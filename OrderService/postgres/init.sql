CREATE TABLE IF NOT EXISTS orderDB 
(
	order_uid VARCHAR(64) PRIMARY KEY,
	data JSONB
);