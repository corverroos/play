#!/bin/bash

set -e

echo "go install engine"
cd /Users/corver/core/repos/unsure
go install github.com/corverroos/unsure/engine/engine

echo "go install play"
cd /Users/corver/core/repos/play
go install github.com/corverroos/play/play
