# Continuous Deployment scripts and config for kod Integration Tests

This folder contains scripts to facilitate the execution of kodit integration tests, and fetching of useful
upstream binaries for performing those tests.

You may wish to purloin code from within here to create your own test environments.

We use Concourse as our CD environment, but the use of scripts and environment variables should mean
you can execute any of our scripts from the CD tool of your choice.

Note that the integration tests live within the same repo as the code, as the code cannot be considered
complete and ready for release without the itnegration tests being executed. They thus have exactly the same
lifecycle and belong in the same Git repo. (Monorepo pattern).
