module github.com/open-ness/EMCO/src/orchestrator

require (
        github.com/open-ness/EMCO/src/rsync v0.0.0-00010101000000-000000000000
)

replace (
        github.com/open-ness/EMCO/src/rsync => ../../../../../rsync
)

go 1.17
