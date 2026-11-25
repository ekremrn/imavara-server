CREATE TABLE generations (
    browser_id uuid NOT NULL REFERENCES browsers,
    operation_key text NOT NULL,
    request_hash bytea NOT NULL,
    status text NOT NULL CHECK (status IN ('reserved','submitted','delivered','failed')),
    asset_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY(browser_id, operation_key),
    CHECK ((status = 'delivered') = (asset_id IS NOT NULL))
);
