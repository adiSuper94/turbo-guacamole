CREATE EXTENSION "uuid-ossp";

CREATE TABLE "chat_rooms" (
  "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "name" varchar(128),
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE TABLE "members" (
  "chat_room_id" uuid,
  "user_id" uuid
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "username" varchar(64),
  "email_id" varchar(128),
  "created_at" timestamp,
  "modified_at" timestamp
);

CREATE TABLE "messages" (
  "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
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
