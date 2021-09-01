CREATE TABLE "advertisement" (
    "id" INT8 NOT NULL,
    "title" VARCHAR(200) NOT NULL,
    "content" TEXT,
    "title_tokens" TSVECTOR,
    "content_tokens" TSVECTOR,
    "thumb_url" VARCHAR(200),
    "tags" TEXT,
    "updated_at" INT4,
    "image_urls" TEXT,
    PRIMARY KEY ("id")
) WITHOUT OIDS;
