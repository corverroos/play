#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

echo "go install engine"
cd ${DIR}/../unsure
go install github.com/corverroos/unsure/engine/engine

echo "go install arena"
cd ${DIR}/../unsure
go install github.com/corverroos/unsure/arena

echo "go install play"
cd ${DIR}
go install github.com/corverroos/play/play

arena -player=play -player_flags='--index=$INDEX|--count=$COUNT' "$@"
