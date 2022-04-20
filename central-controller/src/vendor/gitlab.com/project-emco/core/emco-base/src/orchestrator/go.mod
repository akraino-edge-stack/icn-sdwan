module gitlab.com/project-emco/core/emco-base/src/orchestrator

require (
        gitlab.com/project-emco/core/emco-base/src/rsync v0.0.0-00010101000000-000000000000
)

replace (
        gitlab.com/project-emco/core/emco-base/src/rsync => ../../../../../../../rsync
)

go 1.17
