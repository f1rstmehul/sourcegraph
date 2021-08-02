import classnames from 'classnames'
import React, { useMemo } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { ViewInsightProviderSourceType } from '../../../../core/backend/types'
import { SearchBackendBasedInsight } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight/use-delete-insight'
import { useParallelRequests } from '../../../../hooks/use-parallel-request'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'
import { InsightErrorContent } from '../insight-card/components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from '../insight-card/components/insight-loading-content/InsightLoadingContent'
import { getInsightViewIcon, InsightContentCard } from '../insight-card/InsightContentCard'

import { getBackendInsightById } from './utils/fetch-backend-insight-by-id'

interface BackendInsightProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'>,
        React.HTMLAttributes<HTMLElement> {
    insight: SearchBackendBasedInsight
}

/**
 * Renders BE search based insight. Fetches insight data by gql api handler.
 */
export const BackendInsight: React.FunctionComponent<BackendInsightProps> = props => {
    const { telemetryService, insight, platformContext, settingsCascade, ...otherProps } = props

    const { data, loading, error } = useParallelRequests(
        useMemo(() => () => getBackendInsightById(insight.id), [insight.id])
    )

    const { loading: isDeleting, delete: handleDelete } = useDeleteInsight({
        settingsCascade,
        platformContext,
    })

    return (
        <InsightContentCard
            telemetryService={telemetryService}
            hasContextMenu={true}
            insight={{ id: insight.id, view: data?.view }}
            onDelete={handleDelete}
            {...otherProps}
            className={classnames('be-insight-card', otherProps.className)}
        >
            {loading || isDeleting ? (
                <InsightLoadingContent
                    text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    subTitle={insight.id}
                    icon={getInsightViewIcon(ViewInsightProviderSourceType.Backend)}
                />
            ) : isErrorLike(error) ? (
                <InsightErrorContent
                    error={error}
                    title={insight.id}
                    icon={getInsightViewIcon(ViewInsightProviderSourceType.Backend)}
                />
            ) : (
                data && (
                    <InsightViewContent
                        telemetryService={telemetryService}
                        viewContent={data.view.content}
                        viewID={insight.id}
                        containerClassName="be-insight-card"
                    />
                )
            )}
        </InsightContentCard>
    )
}