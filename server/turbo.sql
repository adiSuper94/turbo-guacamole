CREATE TABLE "chat_rooms" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" varchar(128) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "members" (
  "chat_room_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY ("chat_room_id", "user_id")
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "username" varchar(64) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "messages" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "body" text NOT NULL,
  "chat_room_id" uuid NOT NULL,
  "sender_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "message_deliveries" (
  "message_id" uuid NOT NULL,
  "chat_room_id" uuid NOT NULL,
  "recipient_id" uuid NOT NULL,
  "delivered" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY ("message_id", "chat_room_id", "recipient_id")
);

CREATE UNIQUE INDEX ON "users" ("username");

CREATE UNIQUE INDEX ON "messages" ("id", "chat_room_id");

COMMENT ON COLUMN "messages"."body" IS 'Content of the message';

ALTER TABLE "messages" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("sender_id") REFERENCES "users" ("id");

ALTER TABLE "members" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id");

ALTER TABLE "members" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "message_deliveries" ADD FOREIGN KEY ("message_id", "chat_room_id") REFERENCES "messages" ("id", "chat_room_id");

ALTER TABLE "message_deliveries" ADD FOREIGN KEY ("chat_room_id", "recipient_id") REFERENCES "members" ("chat_room_id", "user_id");
