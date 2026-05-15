package api

import (
	"errors"

	"github.com/ossirdian/magistral/saas/internal/service"
	saasadminv1 "github.com/ossirdian/magistral/shared/pkg/saasadmin/v1"
)

var (
	errNilSchoolService = errors.New("school service is nil")
	errNilBranchService = errors.New("branch service is nil")
)

type api struct {
	schoolService service.SchoolService
	branchService service.BranchService
}

func New(schoolService service.SchoolService, branchService service.BranchService) (*api, error) {
	switch {
	case schoolService == nil:
		return nil, errNilSchoolService
	case branchService == nil:
		return nil, errNilBranchService
	}

	return &api{
		schoolService: schoolService,
		branchService: branchService,
	}, nil
}

func NewServer(schoolService service.SchoolService, branchService service.BranchService, opts ...saasadminv1.ServerOption) (*saasadminv1.Server, error) {
	handler, err := New(schoolService, branchService)
	if err != nil {
		return nil, err
	}

	return saasadminv1.NewServer(handler, opts...)
}
