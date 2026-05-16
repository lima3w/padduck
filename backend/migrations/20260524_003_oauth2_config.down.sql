-- +migrate Down

DROP TABLE IF EXISTS oauth2_states;
DROP TABLE IF EXISTS oauth2_configs;