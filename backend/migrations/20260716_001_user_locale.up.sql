-- +migrate Up

-- Persist the user's preferred UI locale (e.g. 'en', 'fr') so it follows
-- them across devices/sessions instead of living only in localStorage.
ALTER TABLE users
    ADD COLUMN locale VARCHAR(10) NOT NULL DEFAULT 'en';
