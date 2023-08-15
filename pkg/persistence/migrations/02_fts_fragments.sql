CREATE TABLE
  if NOT EXISTS "fragment" (
    "id" INTEGER PRIMARY KEY, -- auto increment doesn't work well for synced tables. also, must be int for use in fts rowid
    "e" VARCHAR(255), -- references some other thing in the db. for now, either a thread or a message
    "t" VARCHAR(255), -- what table this belongs to. not quite using sql the way it was intended here
    "a" VARCHAR(255), -- the name of the a that this fragment is for
    "v" TEXT, -- the v of the a that this fragment is for
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

CREATE VIRTUAL TABLE if NOT EXISTS "fragment_fts" USING fts5 (
  "e" UNINDEXED,
  "t" UNINDEXED,
  "a",
  "v",
  "created_at" UNINDEXED,
  content = "fragment",
  content_rowid = "id",
  tokenize = "trigram"
);

CREATE TRIGGER if NOT EXISTS "fragment_ai" AFTER INSERT ON "fragment" BEGIN
INSERT INTO
  "fragment_fts" ("rowid", "e", "t", "a", "v", "created_at")
VALUES
  (
    NEW."id",
    NEW."e",
    NEW."t",
    NEW."a",
    NEW."v",
    NEW."created_at"
  );

END;

CREATE TRIGGER if NOT EXISTS "fragment_ad" AFTER DELETE ON "fragment" BEGIN
INSERT INTO
  "fragment_fts" (
    "fragment_fts",
    "rowid",
    "e",
    "t",
    "a",
    "v",
    "created_at"
  )
VALUES
  (
    'delete',
    OLD."id",
    OLD."e",
    OLD."t",
    OLD."a",
    OLD."v",
    OLD."created_at"
  );

END;

CREATE TRIGGER if NOT EXISTS "fragment_au" AFTER
UPDATE ON "fragment" BEGIN
INSERT INTO
  "fragment_fts" (
    "fragment_fts",
    "rowid",
    "e",
    "t",
    "a",
    "v",
    "created_at"
  )
VALUES
  (
    'delete',
    OLD."id",
    OLD."e",
    OLD."t",
    OLD."a",
    OLD."v",
    OLD."created_at"
  );

INSERT INTO
  "fragment_fts" ("rowid", "e", "t", "a", "v", "created_at")
VALUES
  (
    NEW."id",
    NEW."e",
    NEW."t",
    NEW."a",
    NEW."v",
    NEW."created_at"
  );

END;