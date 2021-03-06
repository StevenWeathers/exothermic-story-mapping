<script>
    import dragula from 'dragula'
    import Sockette from 'sockette'
    import { onMount, onDestroy } from 'svelte'

    import AddGoal from '../components/AddGoal.svelte'
    import PageLayout from '../components/PageLayout.svelte'
    import UserCard from '../components/UserCard.svelte'
    import InviteUser from '../components/InviteUser.svelte'
    import ColumnForm from '../components/ColumnForm.svelte'
    import StoryForm from '../components/StoryForm.svelte'
    import ColorLegendForm from '../components/ColorLegendForm.svelte'
    import PersonasForm from '../components/PersonasForm.svelte'
    import UsersIcon from '../components/icons/UsersIcon.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import EditIcon from '../components/icons/EditIcon.svelte'
    import DownCarrotIcon from '../components/icons/DownCarrotIcon.svelte'
    import CommentIcon from '../components/icons/CommentIcon.svelte'
    import DeleteStoryboard from '../components/DeleteStoryboard.svelte'
    import { appRoutes, PathPrefix } from '../config'
    import { user } from '../stores.js'

    export let storyboardId
    export let notifications
    export let router
    export let eventTag

    const { AllowRegistration } = appConfig
    const loginOrRegister = AllowRegistration
        ? appRoutes.register
        : appRoutes.login

    const hostname = window.location.origin
    const socketExtension = window.location.protocol === 'https:' ? 'wss' : 'ws'

    // instantiate dragula, utilizing drop-column as class for the containers
    const drake = dragula({
        isContainer: function(el) {
            return el.classList.contains('drop-column')
        },
    })

    let socketError = false
    let socketReconnecting = false
    let storyboard = {
        owner_id: '',
        goals: [],
        users: [],
        colorLegend: [],
        personas: [],
    }
    let showUsers = false
    let showColorLegend = false
    let showColorLegendForm = false
    let showPersonas = false
    let showPersonasForm = null
    let editColumn = null
    let activeStory = null
    let showDeleteStoryboard = false

    // event handlers
    const addStory = (goalId, columnId) => () => {
        sendSocketEvent(
            'add_story',
            JSON.stringify({
                goalId,
                columnId,
            }),
        )
        eventTag('story_add', 'storyboard', '')
    }

    const deleteStory = storyId => () => {
        sendSocketEvent('delete_story', storyId)
        eventTag('story_delete', 'storyboard', '')
    }

    const addStoryColumn = goalId => () => {
        sendSocketEvent(
            'add_column',
            JSON.stringify({
                goalId,
            }),
        )
        eventTag('column_add', 'storyboard', '')
    }

    const deleteColumn = columnId => () => {
        sendSocketEvent('delete_column', columnId)
        eventTag('column_delete', 'storyboard', '')
        toggleColumnEdit()()
    }

    const changeStoryColor = (storyId, color) => () => {
        sendSocketEvent(
            'update_story_color',
            JSON.stringify({
                storyId,
                color,
            }),
        )
        eventTag('story_edit_color', 'storyboard', color)
    }

    const storyUpdateName = storyId => evt => {
        const name = evt.target.value
        sendSocketEvent(
            'update_story_name',
            JSON.stringify({
                storyId,
                name,
            }),
        )
        eventTag('story_edit_name', 'storyboard', '')
    }

    const storyUpdateContent = storyId => evt => {
        const content = evt.target.value
        sendSocketEvent(
            'update_story_content',
            JSON.stringify({
                storyId,
                content,
            }),
        )
        eventTag('story_edit_content', 'storyboard', '')
    }

    const storyUpdatePoints = storyId => evt => {
        const points = parseInt(evt.target.value, 10)
        console.log(points)
        console.log(evt)
        sendSocketEvent(
            'update_story_points',
            JSON.stringify({
                storyId,
                points,
            }),
        )
        eventTag('story_edit_points', 'storyboard', '')
    }

    const storyUpdateClosed = storyId => closed => {
        sendSocketEvent(
            'update_story_closed',
            JSON.stringify({
                storyId,
                closed,
            }),
        )
        eventTag('story_edit_closed', 'storyboard', '')
    }

    drake.on('drop', function(el, target, source, sibling) {
        const storyId = el.dataset.storyid
        const goalId = target.dataset.goalid
        const columnId = target.dataset.columnid

        // determine what story to place story before in target column
        const placeBefore = sibling ? sibling.dataset.storyid : ''

        sendSocketEvent(
            'move_story',
            JSON.stringify({
                storyId,
                goalId,
                columnId,
                placeBefore,
            }),
        )
        eventTag('story_move', 'storyboard', '')
    })

    const onSocketMessage = function(evt) {
        const parsedEvent = JSON.parse(evt.data)

        switch (parsedEvent.type) {
            case 'init':
                storyboard = JSON.parse(parsedEvent.value)
                eventTag('join', 'storyboard', '')
                break
            case 'user_joined':
                storyboard.users = JSON.parse(parsedEvent.value)
                const joinedUser = storyboard.users.find(
                    w => w.id === parsedEvent.userId,
                )
                notifications.success(`${joinedUser.name} joined.`)
                break
            case 'user_retreated':
                const leftUser = storyboard.users.find(
                    w => w.id === parsedEvent.userId,
                )
                storyboard.users = JSON.parse(parsedEvent.value)

                notifications.danger(`${leftUser.name} retreated.`)
                break
            case 'storyboard_updated':
                storyboard = JSON.parse(parsedEvent.value)
                break
            case 'goal_added':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'goal_revised':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'goal_deleted':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'column_added':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'column_updated':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'story_added':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'story_updated':
                storyboard.goals = JSON.parse(parsedEvent.value)
                if (activeStory) {
                    let activeStoryFound = false
                    for (let goal of storyboard.goals) {
                        for (let column of goal.columns) {
                            for (let story of column.stories) {
                                if (story.id === activeStory.id) {
                                    activeStory = story
                                    break
                                }
                            }
                            if (activeStoryFound) {
                                break
                            }
                        }
                        if (activeStoryFound) {
                            break
                        }
                    }
                }
                break
            case 'story_moved':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'story_deleted':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'personas_updated':
                storyboard.personas = JSON.parse(parsedEvent.value)
                break
            case 'storyboard_conceded':
                // storyboard over, goodbye.
                notifications.warning('Storyboard deleted')
                router.route(appRoutes.storyboards)
                break
            default:
                break
        }
    }

    const ws = new Sockette(
        `${socketExtension}://${window.location.host}${PathPrefix}/api/arena/${storyboardId}`,
        {
            timeout: 2e3,
            maxAttempts: 15,
            onmessage: onSocketMessage,
            onerror: () => {
                socketError = true
                eventTag('socket_error', 'storyboard', '')
            },
            onclose: e => {
                if (e.code === 4004) {
                    eventTag('not_found', 'storyboard', '', () => {
                        router.route(appRoutes.storyboards)
                    })
                } else if (e.code === 4001) {
                    eventTag('socket_unauthorized', 'storyboard', '', () => {
                        user.delete()
                        router.route(`${appRoutes.login}/${storyboardId}`)
                    })
                } else if (e.code === 4003) {
                    eventTag('socket_duplicate', 'storyboard', '', () => {
                        notifications.danger(
                            `Duplicate storyboard session exists for your ID`,
                        )
                        router.route(`${appRoutes.storyboards}`)
                    })
                } else if (e.code === 4002) {
                    eventTag(
                        'storyboard_user_abandoned',
                        'storyboard',
                        '',
                        () => {
                            router.route(appRoutes.storyboards)
                        },
                    )
                } else {
                    socketReconnecting = true
                    eventTag('socket_close', 'storyboard', '')
                }
            },
            onopen: () => {
                socketError = false
                socketReconnecting = false
                eventTag('socket_open', 'storyboard', '')
            },
            onmaximum: () => {
                socketReconnecting = false
                eventTag(
                    'socket_error',
                    'storyboard',
                    'Socket Reconnect Max Reached',
                )
            },
        },
    )

    onDestroy(() => {
        eventTag('leave', 'storyboard', '', () => {
            ws.close()
        })
    })

    const sendSocketEvent = (type, value) => {
        ws.send(
            JSON.stringify({
                type,
                value,
            }),
        )
    }

    function concedeStoryboard() {
        eventTag('concede_storyboard', 'storyboard', '', () => {
            sendSocketEvent('concede_storyboard', '')
        })
    }

    function abandonStoryboard() {
        eventTag('abandon_storyboard', 'storyboard', '', () => {
            sendSocketEvent('abandon_storyboard', '')
        })
    }

    function toggleUsersPanel() {
        showColorLegend = false
        showPersonas = false
        showUsers = !showUsers
        eventTag('show_users', 'storyboard', `show: ${showUsers}`)
    }

    function toggleColorLegend() {
        showUsers = false
        showPersonas = false
        showColorLegend = !showColorLegend
        eventTag('show_colorlegend', 'storyboard', `show: ${showColorLegend}`)
    }

    function togglePersonas() {
        showUsers = false
        showColorLegend = false
        showPersonas = !showPersonas
        eventTag('show_personas', 'storyboard', `show: ${showPersonas}`)
    }

    function toggleColumnEdit(column) {
        return () => {
            editColumn = editColumn != null ? null : column
        }
    }

    function toggleEditLegend() {
        showColorLegend = false
        showColorLegendForm = !showColorLegendForm
        eventTag(
            'show_edit_legend',
            'storyboard',
            `show: ${showColorLegendForm}`,
        )
    }

    const toggleEditPersona = persona => () => {
        showPersonas = false
        showPersonasForm = showPersonasForm != null ? null : persona
        eventTag(
            'show_edit_personas',
            'storyboard',
            `show: ${showPersonasForm}`,
        )
    }

    const toggleDeleteStoryboard = () => {
        showDeleteStoryboard = !showDeleteStoryboard
    }

    let showAddGoal = false
    let reviseGoalId = ''
    let reviseGoalName = ''

    const toggleAddGoal = goalId => () => {
        if (goalId) {
            const goalName = storyboard.goals.find(p => p.id === goalId).name
            reviseGoalId = goalId
            reviseGoalName = goalName
        } else {
            reviseGoalId = ''
            reviseGoalName = ''
        }
        showAddGoal = !showAddGoal
        eventTag('show_goal_add', 'storyboard', `show: ${showAddGoal}`)
    }

    const handleGoalAdd = goalName => {
        sendSocketEvent('add_goal', goalName)
        eventTag('goal_add', 'storyboard', '')
    }

    const handleGoalRevision = updatedGoal => {
        sendSocketEvent('revise_goal', JSON.stringify(updatedGoal))
        eventTag('goal_edit_name', 'storyboard', '')
    }

    const handleGoalDeletion = goalId => () => {
        sendSocketEvent('delete_goal', goalId)
        eventTag('goal_delete', 'storyboard', '')
    }

    const handleColumnRevision = column => {
        sendSocketEvent('revise_column', JSON.stringify(column))
        eventTag('column_revise', 'storyboard', '')
    }

    const handleLegendRevision = legend => {
        sendSocketEvent('revise_color_legend', JSON.stringify(legend))
        eventTag('color_legend_revise', 'storyboard', '')
    }

    const addStoryComment = (storyId, comment) => {
        sendSocketEvent(
            'add_story_comment',
            JSON.stringify({ storyId, comment }),
        )
        eventTag('story_add_comment', 'storyboard', '')
    }

    const handlePersonaAdd = persona => {
        sendSocketEvent('add_persona', JSON.stringify(persona))
        eventTag('persona_add', 'storyboard', '')
    }

    const handlePersonaRevision = persona => {
        sendSocketEvent('revise_persona', JSON.stringify(persona))
        eventTag('persona_revise', 'storyboard', '')
    }

    const handleDeletePersona = personaId => () => {
        sendSocketEvent('delete_persona', personaId)
        eventTag('persona_delete', 'storyboard', '')
    }

    const toggleStoryForm = story => () => {
        activeStory = activeStory != null ? null : story
    }

    onMount(() => {
        if (!$user.id) {
            router.route(`${loginOrRegister}/${storyboardId}`)
        }
    })
</script>

<style>
    /** Manually including Dragula styles, should automate this later */
    :global(.gu-mirror) {
        position: fixed !important;
        margin: 0 !important;
        z-index: 9999 !important;
        opacity: 0.8;
        -ms-filter: 'progid:DXImageTransform.Microsoft.Alpha(Opacity=80)';
        filter: alpha(opacity=80);
    }

    :global(.gu-hide) {
        display: none !important;
    }

    :global(.gu-unselectable) {
        -webkit-user-select: none !important;
        -moz-user-select: none !important;
        -ms-user-select: none !important;
        user-select: none !important;
    }

    :global(.gu-transit) {
        opacity: 0.2;
        -ms-filter: 'progid:DXImageTransform.Microsoft.Alpha(Opacity=20)';
        filter: alpha(opacity=20);
    }

    .story-gray {
        @apply border-gray-400;
    }
    .story-gray:hover {
        @apply border-gray-800;
    }
    .story-red {
        @apply border-red-400;
    }
    .story-red:hover {
        @apply border-red-800;
    }
    .story-orange {
        @apply border-orange-400;
    }
    .story-orange:hover {
        @apply border-orange-800;
    }
    .story-yellow {
        @apply border-yellow-400;
    }
    .story-yellow:hover {
        @apply border-yellow-800;
    }
    .story-green {
        @apply border-green-400;
    }
    .story-green:hover {
        @apply border-green-800;
    }
    .story-teal {
        @apply border-teal-400;
    }
    .story-teal:hover {
        @apply border-teal-800;
    }
    .story-blue {
        @apply border-blue-400;
    }
    .story-blue:hover {
        @apply border-blue-800;
    }
    .story-indigo {
        @apply border-indigo-400;
    }
    .story-indigo:hover {
        @apply border-indigo-800;
    }
    .story-purple {
        @apply border-purple-400;
    }
    .story-purple:hover {
        @apply border-purple-800;
    }
    .story-pink {
        @apply border-pink-400;
    }
    .story-pink:hover {
        @apply border-pink-800;
    }

    .colorcard-gray {
        @apply bg-gray-400;
    }
    .colorcard-red {
        @apply bg-red-400;
    }
    .colorcard-orange {
        @apply bg-orange-400;
    }
    .colorcard-yellow {
        @apply bg-yellow-400;
    }
    .colorcard-green {
        @apply bg-green-400;
    }
    .colorcard-teal {
        @apply bg-teal-400;
    }
    .colorcard-blue {
        @apply bg-blue-400;
    }
    .colorcard-indigo {
        @apply bg-indigo-400;
    }
    .colorcard-purple {
        @apply bg-purple-400;
    }
    .colorcard-pink {
        @apply bg-pink-400;
    }
</style>

<svelte:head>
    <title>Storyboard {storyboard.name} | Exothermic</title>
</svelte:head>

{#if storyboard.name && !socketReconnecting && !socketError}
    <div class="px-6 py-2 bg-white flex flex-wrap">
        <div class="w-1/3">
            <h1 class="text-3xl font-bold leading-tight">{storyboard.name}</h1>
        </div>
        <div class="w-2/3 text-right">
            <div>
                {#if storyboard.owner_id === $user.id}
                    <HollowButton
                        color="green"
                        onClick="{toggleAddGoal()}"
                        additionalClasses="mr-2">
                        Add Goal
                    </HollowButton>
                    <HollowButton
                        color="red"
                        onClick="{toggleDeleteStoryboard}"
                        additionalClasses="mr-2">
                        Delete Storyboard
                    </HollowButton>
                {:else}
                    <HollowButton color="red" onClick="{abandonStoryboard}">
                        Leave Storyboard
                    </HollowButton>
                {/if}
                <div class="inline-block relative">
                    <HollowButton
                        color="indigo"
                        additionalClasses="transition ease-in-out duration-150"
                        onClick="{togglePersonas}">
                        Persona's
                        <DownCarrotIcon additionalClasses="ml-1" />
                    </HollowButton>
                    {#if showPersonas}
                        <div
                            class="origin-top-right absolute right-0 mt-1 w-64
                            rounded-md shadow-lg text-left">
                            <div class="rounded-md bg-white shadow-xs">
                                <ul class="p-2">
                                    {#each storyboard.personas as persona}
                                        <li class="mb-1 w-full">
                                            <div>
                                                <span class="font-bold">
                                                    {persona.name}
                                                </span>
                                                {#if storyboard.owner_id === $user.id}
                                                    &nbsp;|&nbsp;
                                                    <button
                                                        on:click="{toggleEditPersona(persona)}"
                                                        class="text-orange-500
                                                        hover:text-orange-800">
                                                        Edit
                                                    </button>
                                                    &nbsp;|&nbsp;
                                                    <button
                                                        on:click="{handleDeletePersona(persona.id)}"
                                                        class="text-red-500
                                                        hover:text-red-800">
                                                        Delete
                                                    </button>
                                                {/if}
                                            </div>
                                            <span class="text-sm">
                                                {persona.role}
                                            </span>
                                        </li>
                                    {/each}
                                </ul>

                                {#if storyboard.owner_id === $user.id}
                                    <div class="p-2 text-right">
                                        <HollowButton
                                            color="green"
                                            onClick="{toggleEditPersona({
                                                id: '',
                                                name: '',
                                                role: '',
                                                description: '',
                                            })}">
                                            Add Persona
                                        </HollowButton>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/if}
                </div>
                <div class="inline-block relative">
                    <HollowButton
                        color="teal"
                        additionalClasses="transition ease-in-out duration-150"
                        onClick="{toggleColorLegend}">
                        Color Legend
                        <DownCarrotIcon additionalClasses="ml-1" />
                    </HollowButton>
                    {#if showColorLegend}
                        <div
                            class="origin-top-right absolute right-0 mt-1 w-64
                            rounded-md shadow-lg text-left">
                            <div class="rounded-md bg-white shadow-xs">
                                <ul class="p-2">
                                    {#each storyboard.color_legend as color}
                                        <li class="mb-1 flex w-full">
                                            <span
                                                class="p-4 mr-2 inline-block
                                                colorcard-{color.color}"></span>
                                            <span
                                                class="inline-block align-middle
                                                {color.legend === '' ? 'text-gray-300' : 'text-gray-600'}">
                                                {color.legend || 'legend not specified'}
                                            </span>
                                        </li>
                                    {/each}
                                </ul>

                                {#if storyboard.owner_id === $user.id}
                                    <div class="p-2 text-right">
                                        <HollowButton
                                            color="orange"
                                            onClick="{toggleEditLegend}">
                                            Edit Legend
                                        </HollowButton>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/if}
                </div>
                <div class="inline-block relative">
                    <HollowButton
                        color="gray"
                        additionalClasses="transition ease-in-out duration-150"
                        onClick="{toggleUsersPanel}">
                        <UsersIcon
                            additionalClasses="mr-1"
                            height="18"
                            width="18" />
                        Users
                        <DownCarrotIcon additionalClasses="ml-1" />
                    </HollowButton>
                    {#if showUsers}
                        <div
                            class="origin-top-right absolute right-0 mt-1 w-64
                            rounded-md shadow-lg text-left">
                            <div class="rounded-md bg-white shadow-xs">
                                {#each storyboard.users as usr, index (usr.id)}
                                    {#if usr.active}
                                        <UserCard
                                            user="{usr}"
                                            {sendSocketEvent}
                                            showBorder="{index != storyboard.users.length - 1}" />
                                    {/if}
                                {/each}

                                <div class="p-2">
                                    <InviteUser
                                        {hostname}
                                        storyboardId="{storyboard.id}" />
                                </div>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
    {#each storyboard.goals as goal, goalIndex (goal.id)}
        <div data-goalid="{goal.id}">
            <div
                class="flex px-6 py-2 border-b-2 bg-gray-200 border-gray-300 {goalIndex > 0 ? 'border-t-2' : ''}">
                <div class="w-3/4 relative">
                    <div class="inline-block align-middle font-bold">
                        <DownCarrotIcon additionalClasses="mr-1" />
                        {goal.name}
                    </div>
                </div>
                <div class="w-1/4 text-right">
                    {#if storyboard.owner_id === $user.id}
                        <HollowButton
                            color="green"
                            onClick="{addStoryColumn(goal.id)}"
                            btnSize="small">
                            Add Column
                        </HollowButton>
                        <HollowButton
                            color="orange"
                            onClick="{toggleAddGoal(goal.id)}"
                            btnSize="small"
                            additionalClasses="ml-2">
                            Edit
                        </HollowButton>
                        <HollowButton
                            color="red"
                            onClick="{handleGoalDeletion(goal.id)}"
                            btnSize="small"
                            additionalClasses="ml-2">
                            Delete
                        </HollowButton>
                    {/if}
                </div>
            </div>
            <section class="flex px-2" style="overflow-x: scroll">
                {#each goal.columns as goalColumn (goalColumn.id)}
                    <div class="flex-none my-4 mx-2 w-40">
                        <div class="flex-none">
                            <div class="w-full mb-2">
                                <div class="flex">
                                    <span
                                        class="font-bold flex-grow truncate"
                                        title="{goalColumn.name}">
                                        {goalColumn.name}
                                    </span>
                                    <button
                                        on:click="{toggleColumnEdit(goalColumn)}"
                                        class="flex-none font-bold text-xl
                                        border-dashed border-2 border-gray-400
                                        hover:border-green-500 text-gray-600
                                        hover:text-green-500 py-1 px-2"
                                        title="Edit Column">
                                        <EditIcon />
                                    </button>
                                </div>
                            </div>
                            <div class="w-full">
                                <div class="flex">
                                    <button
                                        on:click="{addStory(goal.id, goalColumn.id)}"
                                        class="flex-grow font-bold text-xl py-1
                                        px-2 border-dashed border-2
                                        border-gray-400 hover:border-green-500
                                        text-gray-600 hover:text-green-500"
                                        title="Add Story to Column">
                                        +
                                    </button>

                                </div>
                            </div>
                        </div>
                        <ul
                            class="drop-column w-full"
                            style="min-height: 160px;"
                            data-goalid="{goal.id}"
                            data-columnid="{goalColumn.id}">
                            {#each goalColumn.stories as story (story.id)}
                                <li
                                    class="max-w-xs shadow bg-white border-l-4
                                    story-{story.color} border my-4
                                    cursor-pointer"
                                    style="list-style: none;"
                                    data-goalid="{goal.id}"
                                    data-columnid="{goalColumn.id}"
                                    data-storyid="{story.id}"
                                    on:click="{toggleStoryForm(story)}">
                                    <div>
                                        <div>
                                            <div
                                                class="h-20 p-1 text-sm
                                                overflow-hidden {story.closed ? 'line-through' : ''}"
                                                title="{story.name}">
                                                {story.name}
                                            </div>
                                            <div class="h-8">
                                                <div
                                                    class="flex content-center
                                                    p-1 text-sm">
                                                    <div
                                                        class="w-1/2
                                                        text-gray-600">
                                                        {#if story.comments.length > 0}
                                                            <span
                                                                class="inline-block
                                                                align-middle">
                                                                {story.comments.length}
                                                                <CommentIcon />
                                                            </span>
                                                        {/if}
                                                    </div>
                                                    <div
                                                        class="w-1/2 text-right">
                                                        {#if story.points > 0}
                                                            <span
                                                                class="px-2
                                                                bg-gray-300
                                                                inline-block
                                                                align-middle">
                                                                {story.points}
                                                            </span>
                                                        {/if}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </li>
                            {/each}
                        </ul>
                    </div>
                {/each}
            </section>
        </div>
    {/each}
{:else}
    <PageLayout>
        <div class="flex items-center">
            <div class="flex-1 text-center">
                {#if socketReconnecting}
                    <h1
                        class="text-5xl text-orange-500 leading-tight font-bold">
                        Ooops, reloading Storyboard...
                    </h1>
                {:else if socketError}
                    <h1 class="text-5xl text-red-500 leading-tight font-bold">
                        Error joining storyboard, refresh and try again.
                    </h1>
                {:else}
                    <h1 class="text-5xl text-green-500 leading-tight font-bold">
                        Loading Storyboard...
                    </h1>
                {/if}
            </div>
        </div>
    </PageLayout>
{/if}

{#if showAddGoal}
    <AddGoal
        {handleGoalAdd}
        toggleAddGoal="{toggleAddGoal()}"
        {handleGoalRevision}
        goalId="{reviseGoalId}"
        goalName="{reviseGoalName}" />
{/if}

{#if editColumn}
    <ColumnForm
        {handleColumnRevision}
        toggleColumnEdit="{toggleColumnEdit()}"
        column="{editColumn}"
        {deleteColumn} />
{/if}

{#if activeStory}
    <StoryForm
        toggleStoryForm="{toggleStoryForm()}"
        story="{activeStory}"
        changeColor="{changeStoryColor}"
        updateContent="{storyUpdateContent}"
        {deleteStory}
        updateName="{storyUpdateName}"
        updatePoints="{storyUpdatePoints}"
        updateClosed="{storyUpdateClosed}"
        colorLegend="{storyboard.color_legend}"
        addComment="{addStoryComment}"
        users="{storyboard.users}" />
{/if}

{#if showColorLegendForm}
    <ColorLegendForm
        {handleLegendRevision}
        {toggleEditLegend}
        colorLegend="{storyboard.color_legend}" />
{/if}

{#if showPersonasForm}
    <PersonasForm
        toggleEditPersona="{toggleEditPersona()}"
        persona="{showPersonasForm}"
        {handlePersonaAdd}
        {handlePersonaRevision} />
{/if}

{#if showDeleteStoryboard}
    <DeleteStoryboard
        toggleDelete="{toggleDeleteStoryboard}"
        handleDelete="{concedeStoryboard}" />
{/if}
