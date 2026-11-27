-- +migrate Down
DROP TABLE IF EXISTS secret_versions;
DROP TABLE IF EXISTS secrets;
