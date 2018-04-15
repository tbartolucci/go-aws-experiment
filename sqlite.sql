CREATE TABLE "users" ("id" integer primary key autoincrement,"email" varchar(255),"username" varchar(255),"password" varchar(255),"full_name" varchar(255) );

CREATE TABLE "comments" ("id" integer primary key autoincrement,"user_id" integer,"photo_id" integer,"text" varchar(255),"created_at" datetime );

CREATE TABLE "photos" ("id" integer primary key autoincrement,"user_id" integer,"filename" varchar(255),"caption" varchar(255),"created_at" datetime,"likes" integer );

CREATE TABLE "followers" ("user_id" integer,"following_id" integer );
CREATE UNIQUE INDEX idx_user_following ON "followers"(user_id, following_id);

-- username/test
INSERT INTO "users" ("email","username","password","full_name") VALUES ('jdoe@example.com','jdoe','$2a$10$qamAkp1l4dD4pK4zQlaHJumbhZFdC2ZWNvfw.4aNsql1JbKF3vx2S','John Doe');
INSERT INTO "users" ("email","username","password","full_name") VALUES ('jwest@example.com','jwest','$2a$10$qamAkp1l4dD4pK4zQlaHJumbhZFdC2ZWNvfw.4aNsql1JbKF3vx2S','Joyce West');
INSERT INTO "users" ("email","username","password","full_name") VALUES ('smartin@example.com','smartin','$2a$10$qamAkp1l4dD4pK4zQlaHJumbhZFdC2ZWNvfw.4aNsql1JbKF3vx2S','Shirley Martin');
INSERT INTO "users" ("email","username","password","full_name") VALUES ('lallen@example.com','lallen','$2a$10$qamAkp1l4dD4pK4zQlaHJumbhZFdC2ZWNvfw.4aNsql1JbKF3vx2S','Larry Allen');
INSERT INTO "users" ("email","username","password","full_name") VALUES ('sgibson@example.com','sgibson','$2a$10$qamAkp1l4dD4pK4zQlaHJumbhZFdC2ZWNvfw.4aNsql1JbKF3vx2S','Sarah Gibson');

-- create followers

INSERT INTO "followers" ("user_id","following_id") VALUES (1,2);
INSERT INTO "followers" ("user_id","following_id") VALUES (1,3);
INSERT INTO "followers" ("user_id","following_id") VALUES (2,3);
INSERT INTO "followers" ("user_id","following_id") VALUES (2,4);
INSERT INTO "followers" ("user_id","following_id") VALUES (3,4);
INSERT INTO "followers" ("user_id","following_id") VALUES (3,5);
INSERT INTO "followers" ("user_id","following_id") VALUES (4,5);
INSERT INTO "followers" ("user_id","following_id") VALUES (4,1);
INSERT INTO "followers" ("user_id","following_id") VALUES (5,1);
INSERT INTO "followers" ("user_id","following_id") VALUES (5,2);
