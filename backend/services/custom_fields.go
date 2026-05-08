package services

import (
	"context"
	"fmt"

	"ipam-next/models"
	"ipam-next/repository"
)

var validEntityTypes = map[string]bool{
	"subnet": true, "ip_address": true, "device": true,
}

var validFieldTypes = map[string]bool{
	"text": true, "number": true, "textarea": true, "dropdown": true,
	"checkbox": true, "date": true, "url": true, "email": true,
}

func (s *Service) ListCustomFieldDefinitions(ctx context.Context, entityType string) ([]*models.CustomFieldDefinition, error) {
	if entityType != "" && !validEntityTypes[entityType] {
		return nil, fmt.Errorf("invalid entity_type: must be subnet, ip_address, or device")
	}
	return s.repository.ListCustomFieldDefinitions(ctx, entityType)
}

func (s *Service) CreateCustomFieldDefinition(ctx context.Context, p *repository.CustomFieldDefinitionParams) (*models.CustomFieldDefinition, error) {
	if p.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}
	if !validEntityTypes[p.EntityType] {
		return nil, fmt.Errorf("invalid entity_type: must be subnet, ip_address, or device")
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.Label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if !validFieldTypes[p.FieldType] {
		return nil, fmt.Errorf("invalid field_type: must be text, number, textarea, dropdown, checkbox, date, url, or email")
	}
	return s.repository.CreateCustomFieldDefinition(ctx, p)
}

func (s *Service) GetCustomFieldDefinition(ctx context.Context, id int64) (*models.CustomFieldDefinition, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repository.GetCustomFieldDefinition(ctx, id)
}

func (s *Service) UpdateCustomFieldDefinition(ctx context.Context, id int64, p *repository.CustomFieldDefinitionParams) (*models.CustomFieldDefinition, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if p.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}
	if !validEntityTypes[p.EntityType] {
		return nil, fmt.Errorf("invalid entity_type: must be subnet, ip_address, or device")
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.Label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if !validFieldTypes[p.FieldType] {
		return nil, fmt.Errorf("invalid field_type: must be text, number, textarea, dropdown, checkbox, date, url, or email")
	}
	return s.repository.UpdateCustomFieldDefinition(ctx, id, p)
}

func (s *Service) DeleteCustomFieldDefinition(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repository.DeleteCustomFieldDefinition(ctx, id)
}

func (s *Service) ReorderCustomFieldDefinitions(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids is required")
	}
	return s.repository.ReorderCustomFieldDefinitions(ctx, ids)
}

func (s *Service) GetCustomFieldValues(ctx context.Context, entityType string, entityID int64) (map[string]*string, error) {
	if entityID <= 0 {
		return nil, fmt.Errorf("invalid entity ID")
	}
	return s.repository.GetCustomFieldValues(ctx, entityType, entityID)
}

func (s *Service) SetCustomFieldValues(ctx context.Context, entityType string, entityID int64, values map[string]*string) error {
	if entityID <= 0 {
		return fmt.Errorf("invalid entity ID")
	}

	defs, err := s.repository.ListCustomFieldDefinitions(ctx, entityType)
	if err != nil {
		return err
	}

	for _, def := range defs {
		if def.IsRequired {
			val, ok := values[def.Name]
			if !ok || val == nil || *val == "" {
				return fmt.Errorf("field %q is required", def.Name)
			}
		}
	}

	return s.repository.SetCustomFieldValues(ctx, entityType, entityID, defs, values)
}
