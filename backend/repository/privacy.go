package repository

import (
	"context"

	"padduck/models"
)

// ListPrivacyVersions returns all privacy policy versions ordered by created_at DESC.
func (r *Repository) ListPrivacyVersions(ctx context.Context) ([]*models.PrivacyPolicyVersion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, version, effective_date::text, summary, created_at
		FROM privacy_policy_versions
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*models.PrivacyPolicyVersion
	for rows.Next() {
		v := &models.PrivacyPolicyVersion{}
		if err := rows.Scan(&v.ID, &v.Version, &v.EffectiveDate, &v.Summary, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// CreatePrivacyVersion inserts a new privacy policy version record.
func (r *Repository) CreatePrivacyVersion(ctx context.Context, version, effectiveDate string, summary *string) (*models.PrivacyPolicyVersion, error) {
	v := &models.PrivacyPolicyVersion{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO privacy_policy_versions (version, effective_date, summary)
		VALUES ($1, $2::date, $3)
		RETURNING id, version, effective_date::text, summary, created_at`,
		version, effectiveDate, summary,
	).Scan(&v.ID, &v.Version, &v.EffectiveDate, &v.Summary, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ListUserConsentStatus returns consent status for all users (admin report).
func (r *Repository) ListUserConsentStatus(ctx context.Context) ([]*models.UserConsentStatus, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.username, u.email, u.privacy_accepted_at, u.privacy_accepted_version
		FROM users u
		ORDER BY u.username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*models.UserConsentStatus
	for rows.Next() {
		s := &models.UserConsentStatus{}
		if err := rows.Scan(&s.UserID, &s.Username, &s.Email, &s.PrivacyAcceptedAt, &s.PrivacyAcceptedVer); err != nil {
			return nil, err
		}
		s.HasConsent = s.PrivacyAcceptedAt != nil
		statuses = append(statuses, s)
	}
	return statuses, rows.Err()
}
