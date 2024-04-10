CREATE TABLE "chat_rooms" (
  "id" bigserial PRIMARY KEY,
  "name" varchar(128),
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE TABLE "members" (
  "chat_room_id" bigint,
  "user_id" bigint
);

CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar(64),
  "email_id" varchar(128),
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE TABLE "messages" (
  "id" bigserial PRIMARY KEY,
  "body" text,
  "chat_room_id" bigint,
  "sender_id" bigint,
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE UNIQUE INDEX ON "members" ("chat_room_id", "user_id");

COMMENT ON COLUMN "messages"."body" IS 'Content of the message';

ALTER TABLE "messages" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("sender_id") REFERENCES "users" ("id");

ALTER TABLE "members" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id");

ALTER TABLE "members" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
