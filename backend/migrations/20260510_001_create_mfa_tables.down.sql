-- +migrate Down

DROP TABLE IF EXISTS mfa_challenges;
DROP TABLE IF EXISTS user_backup_codes;
DROP TABLE IF EXISTS user_totp_secrets;
DROP TABLE IF EXISTS user_mfa_settings;
