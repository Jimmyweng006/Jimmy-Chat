CREATE TABLE "messages" (
  "id" bigserial PRIMARY KEY,
  "room_id" bigint NOT NULL,
  "reply_message_id" bigint,
  "sender_id" bigint NOT NULL,
  "message_text" text NOT NULL,
  "modified_at" timestamptz NOT NULL DEFAULT (now()),
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "rooms" (
  "id" bigserial PRIMARY KEY,
  "room_name" varchar NOT NULL,
  "created_by" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_room" (
  "user_id" bigint NOT NULL,
  "room_id" bigint NOT NULL,
  PRIMARY KEY ("user_id", "room_id")
);

CREATE INDEX ON "messages" ("id");

CREATE INDEX ON "messages" ("reply_message_id");

CREATE INDEX ON "messages" ("sender_id");

CREATE INDEX ON "rooms" ("room_name");

CREATE INDEX ON "users" ("username");

CREATE INDEX ON "user_room" ("user_id");

CREATE INDEX ON "user_room" ("room_id");

CREATE INDEX ON "user_room" ("user_id", "room_id");

ALTER TABLE "messages" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("reply_message_id") REFERENCES "messages" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("sender_id") REFERENCES "users" ("id");

ALTER TABLE "rooms" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "user_room" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "user_room" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id");

-- defalut value for tables

-- users
INSERT INTO public.users (username,"password",created_at) VALUES
	 ('test1','$2a$10$lc7H443Zca.89milFvWoButsFT/CgRfS.OBP6Xe5tJlvX6THBmicm',now()),
	 ('test2','$2a$10$lc7H443Zca.89milFvWoButsFT/CgRfS.OBP6Xe5tJlvX6THBmicm',now()),
	 ('test3','$2a$10$tNgN.0zzMnfYu7CQePxQe.d/Z7/p12su.EdtorKD8FF/miTFVLN6a',now());

-- rooms
INSERT INTO public.rooms (room_name,created_by,created_at) VALUES ('public',1,now());