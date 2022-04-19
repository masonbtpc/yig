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

CREATE SCHEMA yig;
ALTER SCHEMA yig OWNER TO yig;

SET default_tablespace = '';

--
-- Name: buckets; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.buckets (
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


ALTER TABLE yig.buckets OWNER TO yig;

--
-- Name: cluster; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.cluster (
    fsid character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    weight bigint DEFAULT NULL
);


ALTER TABLE yig.cluster OWNER TO yig;

--
-- Name: gc; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.gc (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version decimal(20) DEFAULT NULL,
    location character varying(255) DEFAULT NULL,
    pool character varying(255) DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    status character varying(255) DEFAULT NULL,
    mtime timestamp with time zone DEFAULT NULL,
    part boolean DEFAULT NULL,
    triedtimes bigint DEFAULT NULL
);


ALTER TABLE yig.gc OWNER TO yig;

--
-- Name: gcpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.gcpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version decimal(20) DEFAULT NULL
);


ALTER TABLE yig.gcpart OWNER TO yig;

--
-- Name: lifecycle; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.lifecycle (
    bucketname character varying(255) DEFAULT NULL,
    status character varying(255) DEFAULT NULL
);


ALTER TABLE yig.lifecycle OWNER TO yig;

--
-- Name: multipartpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.multipartpart (
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


ALTER TABLE yig.multipartpart OWNER TO yig;

--
-- Name: multiparts; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.multiparts (
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


ALTER TABLE yig.multiparts OWNER TO yig;

--
-- Name: objectpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.objectpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version decimal(20) DEFAULT NULL
);


ALTER TABLE yig.objectpart OWNER TO yig;

--
-- Name: objects; Type: TABLE; Schema: yig; Owner: yig
--
CREATE TABLE yig.objects (
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

ALTER TABLE yig.objects OWNER TO yig;

--
-- Name: objmap; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.objmap (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    nullvernum bigint DEFAULT NULL
);


ALTER TABLE yig.objmap OWNER TO yig;

--
-- Name: restoreobjectpart; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.restoreobjectpart (
    partnumber bigint DEFAULT NULL,
    size bigint DEFAULT NULL,
    objectid character varying(255) DEFAULT NULL,
    "offset" bigint DEFAULT NULL,
    etag character varying(255) DEFAULT NULL,
    lastmodified timestamp with time zone DEFAULT NULL,
    initializationvector bytea DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version decimal(20) DEFAULT NULL
);


ALTER TABLE yig.restoreobjectpart OWNER TO yig;

--
-- Name: restoreobjects; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.restoreobjects (
    bucketname character varying(255) DEFAULT NULL,
    objectname character varying(255) DEFAULT NULL,
    version decimal(20) DEFAULT NULL,
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


ALTER TABLE yig.restoreobjects OWNER TO yig;

--
-- Name: users; Type: TABLE; Schema: yig; Owner: yig
--

CREATE TABLE yig.users (
    userid character varying(255) DEFAULT NULL,
    bucketname character varying(255) DEFAULT NULL
);

ALTER TABLE yig.users OWNER TO yig;

--
-- Name: buckets idx_16387_primary; Type: CONSTRAINT; Schema: yig; Owner: yig
--

ALTER TABLE ONLY yig.buckets
    ADD CONSTRAINT idx_16387_primary PRIMARY KEY (bucketname);

--
-- Name: idx_16394_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16394_rowkey ON yig.cluster USING btree (fsid, pool);

--
-- Name: idx_16399_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16399_rowkey ON yig.gc USING btree (bucketname, objectname, version);

--
-- Name: idx_16404_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16404_rowkey ON yig.gcpart USING btree (bucketname, objectname, version);

--
-- Name: idx_16414_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16414_rowkey ON yig.multipartpart USING btree (bucketname, objectname, uploadtime);

--
-- Name: idx_16419_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16419_rowkey ON yig.multiparts USING btree (bucketname, objectname, uploadtime);

--
-- Name: idx_16425_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16425_rowkey ON yig.objectpart USING btree (bucketname, objectname, version);

--
-- Name: idx_16430_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16430_rowkey ON yig.objects USING btree (bucketname, name, version);

--
-- Name: idx_16437_objmap; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16437_objmap ON yig.objmap USING btree (bucketname, objectname);

--
-- Name: idx_16442_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE INDEX idx_16442_rowkey ON yig.restoreobjectpart USING btree (bucketname, objectname, version);

--
-- Name: idx_16447_rowkey; Type: INDEX; Schema: yig; Owner: yig
--

CREATE UNIQUE INDEX idx_16447_rowkey ON yig.restoreobjects USING btree (bucketname, objectname, version);
