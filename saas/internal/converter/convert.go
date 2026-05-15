package converter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/saas/internal/model"
	saasadminv1 "github.com/ossirdian/magistral/shared/pkg/saasadmin/v1"
)

func ConvertSchool(item model.School) (saasadminv1.School, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return saasadminv1.School{}, err
	}

	return saasadminv1.School{
		ID:        id,
		Name:      item.Name,
		Status:    saasadminv1.SchoolStatus(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: nilDateTimeFromPtr(item.DeletedAt),
	}, nil
}

func ConvertSchools(items []model.School) ([]saasadminv1.School, error) {
	result := make([]saasadminv1.School, 0, len(items))
	for _, item := range items {
		converted, err := ConvertSchool(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertBranch(item model.Branch) (saasadminv1.Branch, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return saasadminv1.Branch{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return saasadminv1.Branch{}, err
	}

	return saasadminv1.Branch{
		ID:        id,
		SchoolID:  schoolID,
		Name:      item.Name,
		Address:   nilStringFromPtr(item.Address),
		Status:    saasadminv1.BranchStatus(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: nilDateTimeFromPtr(item.DeletedAt),
	}, nil
}

func ConvertBranches(items []model.Branch) ([]saasadminv1.Branch, error) {
	result := make([]saasadminv1.Branch, 0, len(items))
	for _, item := range items {
		converted, err := ConvertBranch(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertSchoolCreateRequest(req *saasadminv1.SchoolCreateRequest) model.SchoolCreate {
	return model.SchoolCreate{
		Name:   req.Name,
		Status: optSchoolStatusToPtr(req.Status),
	}
}

func ConvertSchoolPatchRequest(req *saasadminv1.SchoolPatchRequest) model.SchoolPatch {
	return model.SchoolPatch{
		Name:   optStringToPtr(req.Name),
		Status: optSchoolStatusToPtr(req.Status),
	}
}

func ConvertBranchCreateRequest(req *saasadminv1.BranchCreateRequest) model.BranchCreate {
	return model.BranchCreate{
		SchoolID: req.SchoolID.String(),
		Name:     req.Name,
		Address:  optNilStringToPtr(req.Address),
		Status:   optBranchStatusToPtr(req.Status),
	}
}

func ConvertBranchPatchRequest(req *saasadminv1.BranchPatchRequest) model.BranchPatch {
	return model.BranchPatch{
		SchoolID: optUUIDToPtr(req.SchoolID),
		Name:     optStringToPtr(req.Name),
		Address:  stringFieldFromOptNilString(req.Address),
		Status:   optBranchStatusToPtr(req.Status),
	}
}

func ConvertSchoolFilter(params saasadminv1.ListSchoolsParams) model.SchoolFilter {
	return model.SchoolFilter{
		Status: optSchoolStatusToPtr(params.Status),
		Search: optStringToPtr(params.Search),
		Limit:  params.Limit.Or(0),
		Offset: params.Offset.Or(0),
	}
}

func ConvertBranchFilter(params saasadminv1.ListBranchesParams) model.BranchFilter {
	return model.BranchFilter{
		SchoolID: optUUIDToPtr(params.SchoolID),
		Status:   optBranchStatusToPtr(params.Status),
		Limit:    params.Limit.Or(0),
		Offset:   params.Offset.Or(0),
	}
}

func ConvertMeta(meta model.PageMeta) saasadminv1.Meta {
	return saasadminv1.Meta{
		Limit:  meta.Limit,
		Offset: meta.Offset,
		Count:  meta.Count,
	}
}

func optNilStringToPtr(v saasadminv1.OptNilString) *string {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optUUIDToPtr(v saasadminv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optStringToPtr(v saasadminv1.OptString) *string {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optSchoolStatusToPtr(v saasadminv1.OptSchoolStatus) *model.SchoolStatus {
	if value, ok := v.Get(); ok {
		status := model.SchoolStatus(value)
		return &status
	}
	return nil
}

func optBranchStatusToPtr(v saasadminv1.OptBranchStatus) *model.BranchStatus {
	if value, ok := v.Get(); ok {
		status := model.BranchStatus(value)
		return &status
	}
	return nil
}

func stringFieldFromOptNilString(v saasadminv1.OptNilString) model.Field[string] {
	field := model.Field[string]{Set: v.IsSet()}
	if value, ok := v.Get(); ok {
		field.Value = &value
	}
	return field
}

func nilStringFromPtr(v *string) saasadminv1.NilString {
	var result saasadminv1.NilString
	if v == nil {
		result.SetToNull()
		return result
	}

	result.SetTo(*v)
	return result
}

func nilDateTimeFromPtr(v *time.Time) saasadminv1.NilDateTime {
	var result saasadminv1.NilDateTime
	if v == nil {
		result.SetToNull()
		return result
	}

	result.SetTo(*v)
	return result
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}
