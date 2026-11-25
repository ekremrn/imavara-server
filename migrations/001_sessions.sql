CREATE TABLE browsers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_hash bytea NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE credit_movements (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    browser_id uuid NOT NULL REFERENCES browsers,
    operation_key text NOT NULL,
    reason text NOT NULL,
    amount integer NOT NULL CHECK (amount <> 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (browser_id, operation_key)
);

CREATE TABLE credit_holds (
    browser_id uuid NOT NULL REFERENCES browsers,
    operation_key text NOT NULL,
    amount integer NOT NULL CHECK (amount > 0),
    PRIMARY KEY (browser_id, operation_key)
);
