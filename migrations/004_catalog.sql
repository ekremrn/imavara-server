CREATE TABLE assets (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 browser_id uuid NOT NULL REFERENCES browsers,
 parent_id uuid REFERENCES assets,
 point_id uuid NOT NULL,
 sha256 text NOT NULL,
 object_key text NOT NULL,
 preview_key text NOT NULL,
 mime_type text NOT NULL,
 width int NOT NULL,
 height int NOT NULL,
 origin text NOT NULL CHECK(origin IN ('generated','contribution')),
 status text NOT NULL DEFAULT 'private',
 license_scope text NOT NULL DEFAULT 'private',
 version integer NOT NULL DEFAULT 1,
 caption text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 published_at timestamptz
);
CREATE UNIQUE INDEX active_content ON assets(sha256) WHERE status IN ('accepted','published');
CREATE INDEX assets_owner ON assets(browser_id,created_at DESC);
CREATE TABLE acquisitions(
 browser_id uuid REFERENCES browsers,
 asset_id uuid REFERENCES assets,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(browser_id,asset_id)
);
CREATE TABLE search_returns(
 search_id uuid NOT NULL,
 asset_id uuid REFERENCES assets,
 rank integer NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(search_id,asset_id)
);
CREATE TABLE catalog_jobs(
 asset_id uuid REFERENCES assets,
 version int NOT NULL,
 published boolean NOT NULL DEFAULT false,
 PRIMARY KEY(asset_id,version)
);
