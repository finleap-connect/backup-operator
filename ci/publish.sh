#!/bin/bash
echo $PWD
export PWD=$(pwd)
export PATH="$PATH:$PWD/tools"

# Setup helm command
echo "Setting up helm..."
ln -sf $PWD/tools/helm3 $PWD/tools/helm

set -eu

# Get a private SSH key from the Github secrets. It will
# be used to establish an identity with rights to push to the git repository
# hosting our Helm charts: https://github.com/kubism/charts
echo "$CHART_PUSH_KEY" > ${PWD}/ci/deploy_key
chmod 0400 ${PWD}/ci/deploy_key

# Activate logging of bash commands now that the sensitive stuff is done
set -x

# As chartpress uses git to push to our Helm chart repository, we configure
# git ahead of time to use the identity we decrypted earlier.
export GIT_SSH_COMMAND="ssh -i ${PWD}/ci/deploy_key"

echo "Publishing chart via chartpress..."
if [ "${TRAVIS_TAG:-}" == "" ]; then
    # Using --long, we are ensured to get a build suffix, which ensures we don't
    # build the same tag twice.
    chartpress --skip-build --publish-chart --long
else
    # Setting a tag explicitly enforces a rebuild if this tag had already been
    # built and we wanted to override it.
    chartpress --skip-build --publish-chart --tag "${TRAVIS_TAG}"
fi

# Let us log the changes chartpress did, it should include replacements for
# fields in values.yaml, such as what tag for various images we are using.
echo "Changes from chartpress:"
git --no-pager diff
