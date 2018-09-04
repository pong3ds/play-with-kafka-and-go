#!/bin/bash

message="'$*'"

if [ -z "$message" ]; then
    echo "Please enter commit message using commit-all <your-message-here>"
    exit 0
fi

git add .
git commit -m "$message"
git push