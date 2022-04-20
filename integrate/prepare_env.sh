#!/usr/bin/env bash
DATABASE=$1
DBPASS=$2

function prepare_ceph(){
	docker exec ceph ceph osd pool create tiger 32
	docker exec ceph ceph osd pool create rabbit 32
	docker exec ceph ceph osd pool create turtle 32
}

function prepare_database(){
    case $DATABASE in
        cockroachdb)
            echo "Preparing Yig for CockroachDB metastore"
            # Prepare Cockroach Environment
            # Note: can't set password in insecure mode with cockroachdb; omitted
            docker cp sql/crdb.sql cockroachdb:/cockroach/crdb.sql
            docker exec cockroachdb bash -c "echo 'create user yig;' | cockroach sql --insecure" > /dev/null 2>&1
            docker exec cockroachdb bash -c "echo 'create database yigdb;'| cockroach sql --insecure" > /dev/null 2>&1
            docker exec cockroachdb bash -c "echo 'grant all on database yigdb to yig;'| cockroach sql --insecure" > /dev/null 2>&1 
            #Create CockroachDB     
            docker exec cockroachdb bash -c "cockroach sql -u yig -d yigdb --insecure < /cockroach/crdb.sql" > /dev/null 2>&1
            ;;
        tidb)
            echo "Preparing Yig for TiDB metastore"
            # Prepare TiDB Environment
            docker cp sql/tidb.sql tidb:/tidb.sql
            docker exec tidb apk update > /dev/null 2>&1
            docker exec tidb apk add mysql-client > /dev/null 2>&1
            docker exec tidb mysql -P 4000 -h 127.0.0.1 -e "create user yig identified by '"${DBPASS}"';"
            docker exec tidb mysql -P 4000 -h 127.0.0.1 -e "create database yigdb character set utf8;"
            docker exec tidb mysql -P 4000 -h 127.0.0.1 -e "grant all privileges on yigdb.* to yig;"
            docker exec tidb mysql -P 4000 -h 127.0.0.1 -e "flush privileges;"
            # Create TiDB
            docker exec tidb mysql -u yig -pBl@rgF1ght -P 4000 -h 127.0.0.1 yigdb -e "source /tidb.sql;"
            ;;
        * )
            echo "Unknown database type specified. Please edit the Makefile. Exiting"
            exit
            ;;
    esac
}

function prepare_vault(){
    echo "start init vault transit..."
    docker exec vault vault secrets enable transit
    docker exec vault vault write -f transit/keys/yig
}

function main(){
    echo "creating Ceph pool..."
    prepare_ceph
    echo "creating Database..."
    prepare_database
    echo "creating Vault..."
    prepare_vault
}

main