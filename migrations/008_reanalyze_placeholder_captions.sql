-- Early v1 analyses with punctuation-only captions must pass the current prompt and validator.
WITH pending AS (
 UPDATE assets SET status='accepted', version=version+1
 WHERE status='published' AND btrim(caption) IN ('...', '…', 'A', '')
 RETURNING id,version
)
INSERT INTO catalog_jobs(asset_id,version) SELECT id,version FROM pending ON CONFLICT DO NOTHING;
