ALTER TABLE service_jobs DROP COLUMN IF EXISTS telegram_message_id;
ALTER TABLE service_jobs DROP COLUMN IF EXISTS telegram_chat_id;
DROP TABLE IF EXISTS telegram_events;
