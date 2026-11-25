ALTER TABLE assets ADD COLUMN generation_prompt text NOT NULL DEFAULT '';

WITH queued AS (
    UPDATE assets SET status='pending_decision', license_scope='shared',
        generation_prompt=caption, version=version+1
    WHERE origin='generated' AND status='private'
    RETURNING id, version
)
INSERT INTO catalog_jobs(asset_id,version) SELECT id,version FROM queued;
