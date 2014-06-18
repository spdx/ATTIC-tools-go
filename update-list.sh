#!/bin/bash

LICENCE_LIST_FILE="licence-list.txt"

# In case git submodule is not initialised
git submodule init

# Update all git submodules (incl. the license-list repo which)
git submodule update --recursive

find ./spdx/license-list -type f |
grep -e '.txt$' | 
grep -vi "Updating the SPDX Licenses" |
grep -vi "README" |
sed 's/\.\/spdx\/license-list\///g' |
sed 's/\.txt//g' > $LICENCE_LIST_FILE


