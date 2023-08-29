package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"movie_api/internal/validator"
	"strings"
	"time"
)

var (
	ErrDuplicatePermission = errors.New("permission already exists")
	ErrNonExistPermission  = errors.New("permission doesn't exist")
	ErrNonExistUser        = errors.New("userID doesn't exist")
)

const (
	MoviesRead       = "movies:read"
	MoviesWrite      = "movies:write"
	CommentsRead     = "comments:read"
	CommentsWrite    = "comments:write"
	UsersRead        = "users:read"
	PermissionsRead  = "permissions:read"
	PermissionsWrite = "permissions:write"
)

type Permissions []string

var PermissionList = Permissions{MoviesRead, MoviesWrite, CommentsRead, CommentsWrite, UsersRead, PermissionsRead, PermissionsWrite}

func ValidatePermissions(v *validator.Validator, p Permissions) {
	permissionsNum := len(PermissionList)
	v.Check(len(p) >= 1, "permissions", "must contain at least 1 code")
	v.Check(len(p) <= permissionsNum, "permissions", fmt.Sprintf("must not contain more than %v code", permissionsNum))
	v.Check(validator.Unique(p), "permissions", "must not contain duplicate values")
	for _, code := range p {
		ValidatePermission(v, code)
	}
}

func ValidatePermission(v *validator.Validator, code string) {
	v.Check(validator.In(code, PermissionList...), "permission", "invalid permission code value")
}

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
SELECT array_agg(permissions.code) as permissions
FROM permissions
INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.id = $1`
	var permissions Permissions
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(pq.Array((*[]string)(&permissions)))
	if err != nil {
		return nil, err
	}
	if permissions == nil {
		permissions = Permissions{}
	}
	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
INSERT INTO users_permissions
SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)
ON CONFLICT ON CONSTRAINT users_permissions_pkey 
DO NOTHING;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `violates foreign key constraint "users_permissions_user_id_fkey"`):
			return ErrNonExistUser
			//		case strings.Contains(err.Error(), `violates unique constraint "users_permissions_pkey"`):
			//			return ErrDuplicatePermission
		default:
			return err
		}
	}
	return nil
}

func (m PermissionModel) DeleteForUser(userID int64, codes ...string) error {
	// Calling the Begin() method on the connection pool creates a new sql.Tx
	// object, which represents the in-progress database transaction.
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	query := `
DELETE FROM users_permissions
WHERE user_id = $1
AND permission_id = ANY(SELECT permissions.id FROM permissions WHERE permissions.code = ANY($2))`

	//Call Exec() on the transaction
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := tx.ExecContext(ctx, query, userID, pq.Array(codes))
	if err != nil {
		// If there is any error, we call the tx.Rollback() method on the
		// transaction. This will abort the transaction and no changes will be
		// made to the database.
		tx.Rollback()
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	// If less rows were affected, we know that the users_permissions table didn't contain a record
	// we tried to delete.
	if int(rowsAffected) < len(codes) {
		tx.Rollback()
		return ErrNonExistPermission
	}
	// If there are no errors, the statements in the transaction can be committed
	// to the database with the tx.Commit() method
	err = tx.Commit()
	return err
}
