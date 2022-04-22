#!/usr/bin/env bash

BASEDIR=$(dirname $(pwd))
BUILDDIR=$1
DATABASE=$2
DBPASS=$3

case $DATABASE in
    cockroachdb)
        echo "Building Yig with CockroachDB"
        db_info="postgres://yig:"${DBPASS}"@10.5.0.17:26257/yigdb"
        meta_store="cockroachdb"
        time_format="2006-01-02 15:04:05"
        ;;
    tidb)
        echo "Building Yig with TiDB"
        db_info="yig:"${DBPASS}"@tcp(10.5.0.17:4000)/yigdb"
        meta_store="tidb"
        time_format="2006-01-02T15:04:05.000Z"
        ;;
    * )
        echo "Unknown database type specified. Defaulting to building Yig with CockroachDB"
        db_info="postgres://yig:"${DBPASS}"10.5.0.17:26257/yigdb"
        meta_store="cockroachdb"
        ;;
esac

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sed -i "s|db_info = \"[^\"]*\"|db_info = \"$db_info\"|" $PWD/yigconf/yig.toml
    sed -i "s|meta_store = \"[^\"]*\"|meta_store = \"$meta_store\"|" $PWD/yigconf/yig.toml
    sed -i "s|time_format = \"[^\"]*\"|time_format = \"$time_format\"|" $PWD/yigconf/yig.toml
elif [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|db_info = \"[^\"]*\"|db_info = \"$db_info\"|" $PWD/yigconf/yig.toml
    sed -i '' "s|meta_store = \"[^\"]*\"|meta_store = \"$meta_store\"|" $PWD/yigconf/yig.toml
    sed -i '' "s|time_format = \"[^\"]*\"|time_format = \"$time_format\"|" $PWD/yigconf/yig.toml
fi

#sudo docker run --rm -v ${BASEDIR}:${BUILDDIR} -w ${BUILDDIR} journeymidnight/yig bash -c 'make build_internal'