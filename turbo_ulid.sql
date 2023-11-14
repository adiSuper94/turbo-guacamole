CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION generate_ulid() RETURNS uuid
    AS $$
        SELECT (lpad(to_hex(floor(extract(epoch FROM clock_timestamp()) * 1000)::bigint), 12, '0') || encode(gen_random_bytes(10), 'hex'))::uuid;
    $$ LANGUAGE SQL;

SELECT generate_ulid();

CREATE TABLE "chat_rooms" (
  "id" uuid PRIMARY KEY DEFAULT generate_ulid(),
  "name" varchar(128),
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE TABLE "members" (
  "chat_room_id" uuid,
  "user_id" uuid
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT generate_ulid(),
  "username" varchar(64),
  "email_id" varchar(128),
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE TABLE "messages" (
  "id" uuid PRIMARY KEY DEFAULT generate_ulid(),
  "body" text,
  "chat_room_id" uuid,
  "sender_id" uuid,
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE UNIQUE INDEX ON "members" ("chat_room_id", "user_id");

COMMENT ON COLUMN "messages"."body" IS 'Content of the message';

ALTER TABLE "messages" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("sender_id") REFERENCES "users" ("id");

ALTER TABLE "members" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id");

ALTER TABLE "members" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
