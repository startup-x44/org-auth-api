package service

import "errors"

// Role-related errors
var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleNotFoundInOrg      = errors.New("role not found in organization")
	ErrCannotUpdateSystemRole = errors.New("cannot update system role")
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")
	ErrRoleHasActiveMembers   = errors.New("cannot delete role with active members")
	ErrRoleNameAlreadyExists  = errors.New("role name already exists in organization")
)

// Permission-related errors
var (
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrCannotModifySystemPerms = errors.New("cannot modify system role permissions")
	ErrCannotUpdateSystemPerm  = errors.New("system permissions cannot be updated")
	ErrCannotDeleteSystemPerm  = errors.New("system permissions cannot be deleted")
	ErrPermissionInUse         = errors.New("permission is currently assigned to roles and cannot be deleted")
	ErrSomePermissionsNotFound = errors.New("some permissions not found")
)

// Organization-related errors
var (
	ErrInsufficientPermission = errors.New("insufficient permissions")
)

// General errors
var (
	ErrInvalidUUID = errors.New("invalid UUID format")
	ErrInvalidData = errors.New("invalid request data")
)
