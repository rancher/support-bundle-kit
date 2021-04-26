#!/bin/bash
set -ex

exec tini -- "${@}"
