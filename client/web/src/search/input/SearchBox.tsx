import classNames from 'classnames'
import React from 'react'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchContextInputProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { VersionContextDropdown } from '../../nav/VersionContextDropdown'
import { VersionContext } from '../../schema/site.schema'
import { QueryState, submitSearch } from '../helpers'

import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import styles from './SearchBox.module.scss'
import { SearchButton } from './SearchButton'
import { SearchContextDropdown } from './SearchContextDropdown'
import { Toggles, TogglesProps } from './toggles/Toggles'

export interface SearchBoxProps
    extends Omit<TogglesProps, 'navbarSearchQuery'>,
        ThemeProps,
        SearchContextInputProps,
        TelemetryProps,
        SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean // significant for query suggestions
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    onFocus?: () => void
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut
    submitSearchOnSearchContextChange?: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined

    /** Whether globbing is enabled for filters. */
    globbing: boolean

    /** Whether comments are parsed and highlighted */
    interpretComments?: boolean

    /** Don't show the version contexts dropdown. */
    hideVersionContexts?: boolean

    /** Don't show search help button */
    hideHelpButton?: boolean
}

export const SearchBox: React.FunctionComponent<SearchBoxProps> = props => {
    const { queryState } = props

    return (
        <div className={classNames(styles.searchBox, props.hideHelpButton ? styles.searchBoxShadow : null)}>
            {!props.hideVersionContexts && (
                <VersionContextDropdown
                    history={props.history}
                    caseSensitive={props.caseSensitive}
                    patternType={props.patternType}
                    navbarSearchQuery={queryState.query}
                    versionContext={props.versionContext}
                    setVersionContext={props.setVersionContext}
                    availableVersionContexts={props.availableVersionContexts}
                    selectedSearchContextSpec={props.selectedSearchContextSpec}
                    className={styles.searchBoxVersionContextDropdown}
                />
            )}
            <div className={classNames(styles.searchBoxBackgroundContainer, 'flex-shrink-past-contents')}>
                {props.searchContextsEnabled && props.showSearchContext && (
                    <>
                        <SearchContextDropdown
                            {...props}
                            query={queryState.query}
                            submitSearch={submitSearch}
                            className={styles.searchBoxContextDropdown}
                        />
                        <div className={styles.searchBoxSeparator} />
                    </>
                )}
                <div className={classNames(styles.searchBoxFocusContainer, 'flex-shrink-past-contents')}>
                    <LazyMonacoQueryInput {...props} className={styles.searchBoxInput} />
                    <Toggles {...props} navbarSearchQuery={queryState.query} className={styles.searchBoxToggles} />
                </div>
            </div>
            <SearchButton hideHelpButton={props.hideHelpButton} className={styles.searchBoxButton} />
        </div>
    )
}
