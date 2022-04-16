function prepare_ceph(){
	docker exec ceph ceph osd pool create tiger 32
	docker exec ceph ceph osd pool create rabbit 32
	docker exec ceph ceph osd pool create turtle 32
}

function prepare_tidb(){
    docker cp yig.sql tidb:/yig.sql
    docker exec tidb apk update > /dev/null 2>&1
    docker exec tidb apk add mysql-client > /dev/null 2>&1
}

function create_tidb(){
	docker exec tidb mysql -P 4000 -h 127.0.0.1 -e "create database yig character set utf8;"
    docker exec tidb mysql -P 4000 -h 127.0.0.1 -e "use yig;source /yig.sql;"
}

function prepare_cockroachdb() {
    docker cp yig_pg.sql cockroachdb:/cockroach/yig_pg.sql
    docker exec cockroachdb bash -c "echo 'create user yig;' | cockroach sql --insecure" > /dev/null 2>&1
    docker exec cockroachdb bash -c "echo 'create database yig;'| cockroach sql --insecure" > /dev/null 2>&1
    docker exec cockroachdb bash -c "echo 'grant all on database yig to yig;'| cockroach sql --insecure" > /dev/null 2>&1 
}

function create_cockroachdb() {
    docker exec cockroachdb bash -c "cockroach sql -u yig -d yig --insecure < /cockroach/yig_pg.sql" > /dev/null 2>&1
}

function prepare_vault(){
    echo "start init vault transit..."
    docker exec vault vault secrets enable transit
    docker exec vault vault write -f transit/keys/yig
}

echo "creating Ceph pool..."
prepare_ceph
echo "preparing TiDB db..."
prepare_tidb
echo "creating TiDB db..."
create_tidb
echo "preparing CockroachDB db..."
prepare_cockroachdb
echo "creating CockroachDB db..."
create_cockroachdb
echo "creating Vault..."
prepare_vault