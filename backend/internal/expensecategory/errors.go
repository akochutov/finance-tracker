package expensecategory

import "errors"

const (
	foreignKeyViolation = "23503"
	uniqueViolation     = "23505"
)

var (
	ErrGroupNotFound     = errors.New("expense group not found")
	ErrCategoryNotFound  = errors.New("expense category not found")
	ErrGroupNameTaken    = errors.New("expense group with this name already exists")
	ErrCategoryNameTaken = errors.New("expense category with this name already exists in the group")
	ErrGroupInactive     = errors.New("expense group is inactive")
	ErrGroupNotEmpty     = errors.New("expense group has active categories")
	ErrNameRequired      = errors.New("name is required")
)
