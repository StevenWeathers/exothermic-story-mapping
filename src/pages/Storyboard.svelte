<script>
    import dragula from 'dragula'
    import Sockette from 'sockette'
    import { onMount, onDestroy } from 'svelte'

    import guid from '../generateGuid.js'
    import AddGoal from '../components/AddGoal.svelte'
    import PageLayout from '../components/PageLayout.svelte'
    import UserCard from '../components/UserCard.svelte'
    import InviteUser from '../components/InviteUser.svelte'
    import UsersIcon from '../components/icons/UsersIcon.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import TimesIcon from '../components/icons/TimesIcon.svelte'
    import TrashIcon from '../components/icons/TrashIcon.svelte'
    import DropperIcon from '../components/icons/DropperIcon.svelte'
    import DownCarrotIcon from '../components/icons/DownCarrotIcon.svelte'

    import { user } from '../stores.js'

    const cardColors = [
        'red',
        'orange',
        'yellow',
        'green',
        'teal',
        'blue',
        'indigo',
        'purple',
        'pink',
    ]

    const defaultColor = 'blue'

    // instantiate dragula, utilizing drop-column as class for the containers
    const drake = dragula({
        isContainer: function(el) {
            return el.classList.contains('drop-column')
        },
    })

    export let storyboardId
    export let notifications
    export let router
    export let eventTag

    const hostname = window.location.origin
    const socketExtension = window.location.protocol === 'https:' ? 'wss' : 'ws'

    let socketError = false
    let socketReconnecting = false
    let storyboard = {
        ownerId: '',
        goals: [],
        users: [],
    }
    let showUsers = false
    let changeColor = ''

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
    }

    const showChangeColor = id => () => (changeColor = id)

    const changeStoryColor = (storyId, color) => () => {
        sendSocketEvent(
            'update_story_color',
            JSON.stringify({
                storyId,
                color,
            }),
        )
        changeColor = ''
        eventTag('story_edit_color', 'storyboard', color)
    }

    const storyUpdateName = storyId => evt => {
        const name = event.target.value
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
        const content = event.target.value
        sendSocketEvent(
            'update_story_content',
            JSON.stringify({
                storyId,
                content,
            }),
        )
        eventTag('story_edit_content', 'storyboard', '')
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
                break
            case 'story_moved':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'story_deleted':
                storyboard.goals = JSON.parse(parsedEvent.value)
                break
            case 'storyboard_conceded':
                // storyboard over, goodbye.
                router.route('/')
                break
            default:
                break
        }
    }

    const ws = new Sockette(
        `${socketExtension}://${window.location.host}/api/arena/${storyboardId}`,
        {
            timeout: 2e3,
            maxAttempts: 15,
            onmessage: onSocketMessage,
            onerror: () => {
                socketError = true
                eventTag('storyboard_error', 'storyboard', 'Socket Error')
            },
            onclose: () => {
                socketReconnecting = true
                eventTag('storyboard_error', 'storyboard', 'Soecket Close')
            },
            onopen: () => {
                socketError = false
                socketReconnecting = false
                eventTag('storyboard_error', 'storyboard', 'Soecket Open')
            },
            onmaximum: () => {
                socketReconnecting = false
                eventTag(
                    'storyboard_error',
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

    function toggleUsersPanel() {
        showUsers = !showUsers
        eventTag('show_users', 'storyboard', `show: ${showUsers}`)
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

    onMount(() => {
        if (!$user.id) {
            router.route(`/register/${storyboardId}`)
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

    .story-red {
        @apply bg-red-100;
        @apply border-red-200;
    }
    .story-orange {
        @apply bg-orange-100;
        @apply border-orange-200;
    }
    .story-yellow {
        @apply bg-yellow-100;
        @apply border-yellow-200;
    }
    .story-green {
        @apply bg-green-100;
        @apply border-green-200;
    }
    .story-teal {
        @apply bg-teal-100;
        @apply border-teal-200;
    }
    .story-blue {
        @apply bg-blue-100;
        @apply border-blue-200;
    }
    .story-indigo {
        @apply bg-indigo-100;
        @apply border-indigo-200;
    }
    .story-purple {
        @apply bg-purple-100;
        @apply border-purple-200;
    }
    .story-pink {
        @apply bg-pink-100;
        @apply border-pink-200;
    }
</style>

<svelte:head>
    <title>Storyboard {storyboard.name} | Exothermic</title>
</svelte:head>

{#if storyboard.name && !socketReconnecting && !socketError}
    <div class="px-6 py-2 bg-white flex flex-wrap">
        <div class="w-2/3">
            <h1 class="text-3xl font-bold leading-tight">{storyboard.name}</h1>
        </div>
        <div class="w-1/3 text-right relative">
            <div>
                {#if storyboard.ownerId === $user.id}
                    <HollowButton
                        color="orange"
                        onClick="{toggleAddGoal()}"
                        additionalClasses="mr-2">
                        Add Goal
                    </HollowButton>
                    <HollowButton
                        color="red"
                        onClick="{concedeStoryboard}"
                        additionalClasses="mr-2">
                        Delete Storyboard
                    </HollowButton>
                {/if}
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
            </div>
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
    {#each storyboard.goals as goal, goalIndex (goal.id)}
        <div data-goalid="{goal.id}">
            <div
                class="flex px-6 py-2 border-b-2 border-gray-300 {goalIndex > 0 ? 'border-t-2' : ''}">
                <div class="w-3/4 relative">
                    <div class="inline-block align-middle font-bold">
                        <DownCarrotIcon additionalClasses="mr-1" />
                        {goal.name}
                    </div>
                </div>
                <div class="w-1/4 text-right">
                    {#if storyboard.ownerId === $user.id}
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
            <section
                class="flex px-2"
                style="overflow-x: scroll; min-height: 260px">
                {#each goal.columns as goalColumn (goalColumn.id)}
                    <div class="flex-none my-4 mx-2" style="width: 260px">
                        <div class="flex">
                            <button
                                on:click="{addStory(goal.id, goalColumn.id)}"
                                class="flex-grow font-bold text-xl bg-gray-300
                                py-1 px-2 mr-1">
                                +
                            </button>
                            <button
                                on:click="{deleteColumn(goalColumn.id)}"
                                class="flex-none font-bold text-xl bg-gray-300
                                py-1 px-2">
                                <TrashIcon />
                            </button>
                        </div>
                        <ul
                            class="drop-column list-reset w-full min-h-full"
                            data-goalid="{goal.id}"
                            data-columnid="{goalColumn.id}">
                            {#each goalColumn.stories as story (story.id)}
                                <li
                                    class="max-w-xs shadow story-{story.color}
                                    border my-4 list-reset"
                                    data-goalid="{goal.id}"
                                    data-columnid="{goalColumn.id}"
                                    data-storyid="{story.id}">
                                    <div class="p-2">
                                        <div class="mb-2 relative flex">
                                            <button
                                                on:click="{deleteStory(story.id)}">
                                                <TimesIcon
                                                    color="{story.color}" />
                                            </button>
                                            <input
                                                type="text"
                                                value="{story.name}"
                                                on:change="{storyUpdateName(story.id)}"
                                                class="inline-block font-bold
                                                text-l bg-transparent mx-2
                                                w-full" />
                                            <div
                                                class="inline-block align-middle
                                                text-right">
                                                <button
                                                    on:click="{showChangeColor(story.id)}">
                                                    <DropperIcon
                                                        color="{story.color}" />
                                                </button>
                                                {#if changeColor === story.id}
                                                    <div
                                                        class="shadow border
                                                        bg-white absolute
                                                        right-0 top-0">
                                                        {#each cardColors as color}
                                                            <button
                                                                on:click="{changeStoryColor(story.id, color)}"
                                                                class="p-4
                                                                hover:bg-{color}-200
                                                                bg-{color}-100"></button>
                                                        {/each}
                                                    </div>
                                                {/if}
                                            </div>
                                        </div>
                                        <textarea
                                            class="w-full h-full bg-transparent
                                            resize-none"
                                            rows="4"
                                            on:change="{storyUpdateContent(story.id)}"
                                            value="{story.content}"></textarea>
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
