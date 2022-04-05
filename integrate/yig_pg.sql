-- Dumped from database version 14.2 (Debian 14.2-1.pgdg110+1)
-- Dumped by pg_dump version 14.2 (Debian 14.2-1.pgdg110+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: yig; Type: SCHEMA; Schema: -; Owner: yig
--
CREATE DATABASE yig;

CREATE USER yig;
GRANT ALL ON DATABASE yig TO yig;

CREATE SCHEMA yig;
ALTER SCHEMA yig OWNER TO yig;

SET default_tablespace = '';

--
-- Name: buckets; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE buckets (
    bucketname character varying(255) DEFAULT ''::character varying NOT NULL,
    acl json,
    cors json,
    logging json DEFAULT 'null'::json NOT NULL,
    lc json,
    uid character varying(255),
    policy json,
    website json,
    encryption json,
    createtime timestamp with time zone,
    usages bigint,
    versioning character varying(255)
);


ALTER TABLE buckets OWNER TO yig;

--
-- Name: cluster; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE cluster (
    fsid character varying(255),
    pool character varying(255),
    weight bigint
);


ALTER TABLE cluster OWNER TO yig;

--
-- Name: gc; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE gc (
    bucketname character varying(255),
    objectname character varying(255),
    version numeric,
    location character varying(255),
    pool character varying(255),
    objectid character varying(255),
    status character varying(255),
    mtime timestamp with time zone,
    part boolean,
    triedtimes bigint
);


ALTER TABLE gc OWNER TO yig;

--
-- Name: gcpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE gcpart (
    partnumber bigint,
    size bigint,
    objectid character varying(255),
    "offset" bigint,
    etag character varying(255),
    lastmodified timestamp with time zone,
    initializationvector bytea,
    bucketname character varying(255),
    objectname character varying(255),
    version numeric
);


ALTER TABLE gcpart OWNER TO yig;

--
-- Name: lifecycle; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE lifecycle (
    bucketname character varying(255),
    status character varying(255)
);


ALTER TABLE lifecycle OWNER TO yig;

--
-- Name: multipartpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE multipartpart (
    partnumber bigint,
    size bigint,
    objectid character varying(255),
    "offset" bigint,
    etag character varying(255),
    lastmodified timestamp with time zone,
    initializationvector bytea,
    bucketname character varying(255),
    objectname character varying(255),
    uploadtime numeric
);


ALTER TABLE multipartpart OWNER TO yig;

--
-- Name: multiparts; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE multiparts (
    bucketname character varying(255),
    objectname character varying(255),
    uploadtime numeric,
    initiatorid character varying(255),
    ownerid character varying(255),
    contenttype character varying(255),
    location character varying(255),
    pool character varying(255),
    acl json,
    sserequest json,
    encryption bytea,
    cipher bytea,
    attrs json,
    storageclass boolean DEFAULT false
);


ALTER TABLE multiparts OWNER TO yig;

--
-- Name: objectpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE objectpart (
    partnumber bigint,
    size bigint,
    objectid character varying(255),
    "offset" bigint,
    etag character varying(255),
    lastmodified timestamp with time zone,
    initializationvector bytea,
    bucketname character varying(255),
    objectname character varying(255),
    version character varying(255)
);


ALTER TABLE objectpart OWNER TO yig;

--
-- Name: objects; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE objects (
    bucketname character varying(255),
    name character varying(255),
    version numeric,
    location character varying(255),
    pool character varying(255),
    ownerid character varying(255),
    size bigint,
    objectid character varying(255),
    lastmodifiedtime timestamp with time zone,
    etag character varying(255),
    contenttype character varying(255),
    customattributes json,
    acl json,
    nullversion boolean,
    deletemarker boolean,
    ssetype character varying(255),
    encryptionkey bytea,
    initializationvector bytea,
    type boolean DEFAULT false,
    storageclass boolean DEFAULT false
);


ALTER TABLE objects OWNER TO yig;

--
-- Name: objmap; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE objmap (
    bucketname character varying(255),
    objectname character varying(255),
    nullvernum bigint
);


ALTER TABLE objmap OWNER TO yig;

--
-- Name: restoreobjectpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE restoreobjectpart (
    partnumber bigint,
    size bigint,
    objectid character varying(255),
    "offset" bigint,
    etag character varying(255),
    lastmodified timestamp with time zone,
    initializationvector bytea,
    bucketname character varying(255),
    objectname character varying(255),
    version numeric
);


ALTER TABLE restoreobjectpart OWNER TO yig;

--
-- Name: restoreobjects; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE restoreobjects (
    bucketname character varying(255),
    objectname character varying(255),
    version numeric,
    status boolean DEFAULT false,
    lifetime smallint DEFAULT '1'::smallint,
    lastmodifiedtime timestamp with time zone,
    location character varying(255),
    pool character varying(255),
    ownerid character varying(255),
    size bigint,
    objectid character varying(255),
    etag character varying(255)
);


ALTER TABLE restoreobjects OWNER TO yig;

--
-- Name: users; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE users (
    userid character varying(255),
    bucketname character varying(255)
);

ALTER TABLE users OWNER TO yig;

--
-- Name: buckets idx_16387_primary; Type: CONSTRAINT; Schema: yig; Owner: yig
--

ALTER TABLE ONLY buckets
    ADD CONSTRAINT idx_16387_primary PRIMARY KEY (bucketname);

--
-- Name: idx_16394_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16394_rowkey ON cluster USING btree (fsid, pool);

--
-- Name: idx_16399_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16399_rowkey ON gc USING btree (bucketname, objectname, version);

--
-- Name: idx_16404_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16404_rowkey ON gcpart USING btree (bucketname, objectname, version);

--
-- Name: idx_16414_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16414_rowkey ON multipartpart USING btree (bucketname, objectname, uploadtime);

--
-- Name: idx_16419_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16419_rowkey ON multiparts USING btree (bucketname, objectname, uploadtime);

--
-- Name: idx_16425_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16425_rowkey ON objectpart USING btree (bucketname, objectname, version);

--
-- Name: idx_16430_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16430_rowkey ON objects USING btree (bucketname, name, version);

--
-- Name: idx_16437_objmap; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16437_objmap ON objmap USING btree (bucketname, objectname);

--
-- Name: idx_16442_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16442_rowkey ON restoreobjectpart USING btree (bucketname, objectname, version);

--
-- Name: idx_16447_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16447_rowkey ON restoreobjects USING btree (bucketname, objectname, version);
