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

CREATE USER yig;
CREATE DATABASE yig;
GRANT ALL ON DATABASE yig TO yig;
CREATE SCHEMA yig;
ALTER SCHEMA yig OWNER TO yig;

SET default_tablespace = '';

--
-- Name: buckets; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE buckets (
    bucketname character varying(255) DEFAULT ''::character varying NOT NULL,
    acl json DEFAULT NULL,
    cors json DEFAULT NULL,
    logging json DEFAULT 'null'::json NOT NULL,
    lc json DEFAULT NULL,
    uid character varying(255) DEFAULT NULL,
    policy json DEFAULT NULL,
    website json DEFAULT NULL,
    encryption json DEFAULT NULL,
    createtime timestamp with time zone DEFAULT NULL,
    usages bigint DEFAULT NULL,
    versioning character varying(255)
);


ALTER TABLE buckets OWNER TO yig;

--
-- Name: cluster; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE cluster (
    fsid character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    weight bigint DEFAULT NULL
);


ALTER TABLE cluster OWNER TO yig;

--
-- Name: gc; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE gc (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version numeric DEFAULT NULL,
    location character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    status character varying(255) DEFAULT NULL,
    mtime timestamp with time zone DEFAULT NULL,
    part smallint DEFAULT NULL,
    triedtimes bigint DEFAULT NULL
);


ALTER TABLE gc OWNER TO yig;

--
-- Name: gcpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE gcpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version numeric DEFAULT NULL
);


ALTER TABLE gcpart OWNER TO yig;

--
-- Name: lifecycle; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE lifecycle (
    bucketname character varying(255) DEFAULT NULL,
    status character varying(255) DEFAULT NULL
);


ALTER TABLE lifecycle OWNER TO yig;

--
-- Name: multipartpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE multipartpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255),
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL, 
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    uploadtime numeric DEFAULT NULL
);


ALTER TABLE multipartpart OWNER TO yig;

--
-- Name: multiparts; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE multiparts (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    uploadtime numeric DEFAULT NULL,
    initiatorid character varying(255) DEFAULT NULL,
    ownerid character varying(255) DEFAULT NULL,
    contenttype character varying(255) DEFAULT NULL,
    location character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    acl json DEFAULT NULL,
    sserequest json DEFAULT NULL,
    encryption bytea DEFAULT NULL,
    cipher bytea DEFAULT NULL,
    attrs json DEFAULT NULL,
    storageclass smallint DEFAULT '0'::smallint 
);


ALTER TABLE multiparts OWNER TO yig;

--
-- Name: objectpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE objectpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version character varying(255) DEFAULT NULL
);


ALTER TABLE objectpart OWNER TO yig;

--
-- Name: objects; Type: TABLE; Schema: yig; Owner: yig
--
CREATE TABLE objects (
    bucketname character varying(255) DEFAULT NULL,
    name character varying(255) DEFAULT NULL,
    version decimal(20) DEFAULT NULL,
    location character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    ownerid character varying(255) DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    lastmodifiedtime timestamp with time zone DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    contenttype character varying(255) DEFAULT NULL,
    customattributes json DEFAULT NULL,
    acl json DEFAULT NULL,
    nullversion boolean DEFAULT NULL,
    deletemarker boolean DEFAULT NULL,
    ssetype character varying(255) DEFAULT NULL,
    encryptionkey bytea DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    type smallint DEFAULT '0'::smallint,
    storageclass smallint DEFAULT '0'::smallint
);

ALTER TABLE objects OWNER TO yig;

--
-- Name: objmap; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE objmap (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    nullvernum bigint DEFAULT NULL
);


ALTER TABLE objmap OWNER TO yig;

--
-- Name: restoreobjectpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE restoreobjectpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version numeric DEFAULT NULL
);


ALTER TABLE restoreobjectpart OWNER TO yig;

--
-- Name: restoreobjects; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE restoreobjects (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version numeric DEFAULT NULL,
    status smallint DEFAULT '0'::smallint,
    lifetime smallint DEFAULT '1'::smallint,
    lastmodifiedtime timestamp with time zone DEFAULT NULL,
    location character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    ownerid character varying(255) DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    etag character varying(255) DEFAULT NULL
);


ALTER TABLE restoreobjects OWNER TO yig;

--
-- Name: users; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE users (
    userid character varying(255) DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL
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
