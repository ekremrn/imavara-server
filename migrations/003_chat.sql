CREATE TABLE conversations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 browser_id uuid NOT NULL REFERENCES browsers,
 title text NOT NULL DEFAULT 'Yeni sohbet',
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX conversations_owner ON conversations(browser_id,created_at DESC);
CREATE TABLE messages (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 conversation_id uuid NOT NULL REFERENCES conversations ON DELETE CASCADE,
 turn_id uuid NOT NULL,
 role text NOT NULL CHECK(role IN ('user','assistant')),
 content text NOT NULL,
 cards jsonb NOT NULL DEFAULT '[]',
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(conversation_id,turn_id,role)
);
