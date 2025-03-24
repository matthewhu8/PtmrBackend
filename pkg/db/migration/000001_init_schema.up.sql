CREATE TYPE role AS ENUM ('employer', 'candidate');
CREATE TYPE application_status AS ENUM ('pending', 'submitted', 'accepted', 'rejected');
CREATE TYPE job_preference AS ENUM ('in person', 'remote', 'open', 'hybrid');
CREATE TYPE education as ENUM ('high school diploma', 'associate', 'bachelor', 'master', 'doctoral');
CREATE TYPE swipe as ENUM('accept', 'reject');
CREATE TABLE "users" (
                         "username" varchar PRIMARY KEY,
                         "email" varchar(255) UNIQUE NOT NULL,
                         "hashed_password" varchar(255) NOT NULL,
                         "role" role NOT NULL,
                         "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "employers" (
                             "id" bigserial PRIMARY KEY,
                             "username" varchar UNIQUE NOT NULL,
                             "business_name" varchar NOT NULL,
                             "business_email" varchar UNIQUE NOT NULL,
                             "business_phone" varchar NOT NULL,
                             "location" varchar NOT NULL,
                             "industry" varchar NOT NULL,
                             "profile_photo" varchar NOT NULL,
                             "business_description" text NOT NULL,
                             "job_listings" varchar(32)[] NOT NULL DEFAULT '{}',
                             "created_at" timestamptz NOT NULL DEFAULT (now()),
                             FOREIGN KEY ("username") REFERENCES "users" ("username") ON DELETE CASCADE
);

CREATE TABLE "candidates" (
                         "id" bigserial PRIMARY KEY,
                         "username" varchar UNIQUE NOT NULL,
                         "full_name" varchar NOT NULL,
                         "phone_number" varchar NOT NULL,
                         "education" education NOT NULL,
                         "location" varchar NOT NULL,
                         "skill_set" varchar(30)[] NOT NULL,
                         "certificates" varchar(30)[] NOT NULL,
                         "industry_of_interest" varchar(20) NOT NULL,
                         "job_preference" job_preference NOT NULL,
                         "time_availability" jsonb NOT NULL,
                         "account_verified" boolean NOT NULL,
                         "resume_file" varchar NOT NULL,
                         "profile_photo" varchar NOT NULL,
                         "description" text NOT NULL,
                         "created_at" timestamptz NOT NULL DEFAULT now(),
                         FOREIGN KEY ("username") REFERENCES "users" ("username") ON DELETE CASCADE
);

CREATE TABLE "past_experiences" (
                                    "id" bigserial PRIMARY KEY,
                                    "candidate_id" bigint NOT NULL,
                                    "industry" varchar NOT NULL,
                                    "employer" varchar NOT NULL,
                                    "job_title" varchar NOT NULL,
                                    "start_date" DATE NOT NULL,
                                    "end_date" DATE NOT NULL,
                                    "present" boolean NOT NULL,
                                    "description" text NOT NULL,
                                    "created_at" timestamptz NOT NULL DEFAULT now(),
                                    FOREIGN KEY ("candidate_id") REFERENCES "candidates" ("id") ON DELETE CASCADE
);

CREATE TABLE candidate_applications (
                                        "candidate_id" bigint NOT NULL,
                                        "employer_id" bigint NOT NULL,
                                        "elasticsearch_doc_id" varchar UNIQUE NOT NULL,
                                        "job_doc_id" varchar UNIQUE NOT NULL,
                                        "application_status" application_status NOT NULL,
                                        "created_at" timestamptz NOT NULL DEFAULT now(),
                                        PRIMARY KEY ("candidate_id", "employer_id"),
                                        FOREIGN KEY ("candidate_id") REFERENCES "candidates" ("id"),
                                        FOREIGN KEY ("employer_id") REFERENCES "employers" ("id")
);

CREATE TABLE employer_applications (
                                        "employer_id" bigint NOT NULL,
                                        "candidate_id" bigint NOT NULL,
                                        "message" text NOT NULL,
                                        "application_status" application_status NOT NULL,
                                        "created_at" timestamptz NOT NULL DEFAULT now(),
                                        PRIMARY KEY ("employer_id", "candidate_id"),
                                        FOREIGN KEY ("candidate_id") REFERENCES "candidates" ("id"),
                                        FOREIGN KEY ("employer_id") REFERENCES "employers" ("id")
);

CREATE TABLE candidate_swipes (
                                  "candidate_id" bigint NOT NULL,
                                  "job_id" varchar NOT NULL,
                                  "swipe" swipe NOT NULL,
                                  "created_at" timestamptz NOT NULL DEFAULT now(),
                                  PRIMARY KEY ("candidate_id", "job_id"),
                                  FOREIGN KEY ("candidate_id") REFERENCES "candidates" ("id")
);

CREATE TABLE employer_swipes (
                                  "employer_id" bigint NOT NULL,
                                  "candidate_id" bigint NOT NULL,
                                  "swipe" swipe NOT NULL,
                                  "created_at" timestamptz NOT NULL DEFAULT now(),
                                  PRIMARY KEY ("employer_id", "candidate_id"),
                                  FOREIGN KEY ("candidate_id") REFERENCES "candidates" ("id"),
                                  FOREIGN KEY ("employer_id") REFERENCES "employers" ("id")
);
