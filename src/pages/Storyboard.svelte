<script>
    import dragula from 'dragula'
    import Sockette from 'sockette'
    import { onMount, onDestroy } from 'svelte'

    import guid from '../generateGuid.js'
    import PageLayout from '../components/PageLayout.svelte'
    import UserCard from '../components/UserCard.svelte'
    import InviteUser from '../components/InviteUser.svelte'
    import UsersIcon from '../components/icons/UsersIcon.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import TimesIcon from '../components/icons/TimesIcon.svelte'
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

    const hostname = window.location.origin
    const socketExtension = window.location.protocol === 'https:' ? 'wss' : 'ws'

    let socketError = false
    let socketReconnecting = false
    let storyboard = {}
    let showUsers = false

    // All the Stateful things...
    let notesColumns = []
    let changeColor = ''

    // event handlers
    const addNote = index => () => {
        const id = guid()

        notesColumns[index].notes[notesColumns[index].notes.length] = {
            id,
            color: defaultColor,
            title: 'Note Title Here',
            content: `Note Content Here ${id}`,
            contentHeight: null,
        }

        sendChanges()
    }

    const deleteNote = (column, index) => () => {
        notesColumns[column].notes.splice(index, 1)
        notesColumns = notesColumns

        sendChanges()
    }

    const addNoteColumn = () => {
        notesColumns[notesColumns.length] = {
            id: guid(),
            notes: [],
        }

        sendChanges()
    }

    const showChangeColor = id => () => (changeColor = id)

    const changeNoteColor = (column, index, color) => () => {
        notesColumns[column].notes[index].color = color
        changeColor = ''

        sendChanges()
    }

    const noteContentEdit = (column, index, type) => event => {
        notesColumns[column].notes[index][type] = event.target.value

        sendChanges()
    }

    drake.on('drop', function(el, target, source, sibling) {
        const noteId = el.dataset.noteid
        const targetColIndex = target.dataset.columnindex
        const sourceColIndex = source.dataset.columnindex

        // determine what order to place note in column
        const finalIndex = sibling
            ? sibling.dataset.index
            : notesColumns[targetColIndex].notes.length

        const note = notesColumns[sourceColIndex].notes.find(
            n => n.id === noteId,
        )

        // remote note from source column
        notesColumns[sourceColIndex].notes.splice(
            notesColumns[sourceColIndex].notes.indexOf(note),
            1,
        )

        // add note to target column
        notesColumns[targetColIndex].notes.splice(finalIndex, 0, note)

        // finally update state
        notesColumns = notesColumns

        sendChanges()
    })

    // catch and update textarea adjusted size in data
    const detectElementMouseEnlargement = (column, index) => ev => {
        const element = ev.target
        const size = { height: element.clientHeight }
        let styleHeight = parseFloat(
            getComputedStyle(element)['height'].replace('px', ''),
        )

        const mouseMoveListener = event => {
            if (element.clientHeight != size.height) {
                let style = getComputedStyle(element)
                styleHeight = parseFloat(style['height'].replace('px', ''))

                size.height = element.clientHeight
            }
        }

        const mouseUpListener = event => {
            window.removeEventListener('mousemove', mouseMoveListener)
            window.removeEventListener('mouseup', mouseUpListener)
            notesColumns[column].notes[index].contentHeight = styleHeight

            sendChanges()
        }

        window.addEventListener('mousemove', mouseMoveListener)
        window.addEventListener('mouseup', mouseUpListener)
    }

    const sendChanges = () => {
        sendSocketEvent('stories_updated', notesColumns)
    }

    const onSocketMessage = function(evt) {
        const parsedEvent = JSON.parse(evt.data)

        switch (parsedEvent.type) {
            case 'init':
                storyboard = JSON.parse(parsedEvent.value)
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
            },
            onclose: () => {
                socketReconnecting = true
            },
            onopen: () => {
                socketError = false
                socketReconnecting = false
            },
            onmaximum: () => {
                socketReconnecting = false
            },
        },
    )

    onDestroy(() => {
        ws.close()
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
        sendSocketEvent('concede_storyboard', '')
    }

    function toggleUsersPanel() {
        showUsers = !showUsers
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

    .note-red {
        @apply bg-red-100;
        @apply border-red-200;
    }
    .note-orange {
        @apply bg-orange-100;
        @apply border-orange-200;
    }
    .note-yellow {
        @apply bg-yellow-100;
        @apply border-yellow-200;
    }
    .note-green {
        @apply bg-green-100;
        @apply border-green-200;
    }
    .note-teal {
        @apply bg-teal-100;
        @apply border-teal-200;
    }
    .note-blue {
        @apply bg-blue-100;
        @apply border-blue-200;
    }
    .note-indigo {
        @apply bg-indigo-100;
        @apply border-indigo-200;
    }
    .note-purple {
        @apply bg-purple-100;
        @apply border-purple-200;
    }
    .note-pink {
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
            <h1 class="text-3xl font-bold leading-tight">
                {storyboard.name}
            </h1>
        </div>
        <div class="w-1/3 text-right relative">
            <div>
                {#if storyboard.ownerId === $user.id}
                    <HollowButton
                        color="red"
                        onClick="{concedeStoryboard}"
                        additionalClasses="mr-2"
                    >
                        Delete Storyboard
                    </HollowButton>
                {/if}
                <HollowButton
                    color="gray"
                    additionalClasses="mr-2 transition ease-in-out duration-150"
                    onClick="{toggleUsersPanel}"
                >
                    <UsersIcon additionalClasses="mr-1" height="18" width="18" />
                    Users
                    <DownCarrotIcon additionalClasses="ml-1" />
                </HollowButton>
            </div>
            {#if showUsers}
            <div class="origin-top-right absolute right-0 mt-1 w-64 rounded-md shadow-lg text-left">
                <div class="rounded-md bg-white shadow-xs">
                    {#each storyboard.users as usr, index (usr.id)}
                        {#if usr.active}
                            <UserCard user="{usr}" {sendSocketEvent} showBorder={index != storyboard.users.length - 1} />
                        {/if}
                    {/each}

                    <div class="p-2">
                        <InviteUser {hostname} storyboardId="{storyboard.id}" />
                    </div>
                </div>
            </div>
            {/if}
        </div>
    </div>
    <div class="h-screen">
        <section
            class="flex items-stretch min-h-full"
            style="overflow-x: scroll">
            {#each notesColumns as noteColumn, index}
                <div class="flex-no-shrink m-3" style="width: 320px">
                    <button
                        on:click="{addNote(index)}"
                        class="w-full font-bold text-3xl bg-gray-300 p-1">
                        +
                    </button>
                    <ul
                        class="drop-column list-reset w-full min-h-full"
                        data-columnid="{noteColumn.id}"
                        data-columnindex="{index}">
                        {#each noteColumn.notes as note, i (note.id)}
                            <li
                                class="max-w-xs shadow note-{note.color} border
                                mt-5 list-reset"
                                data-index="{i}"
                                data-noteid="{note.id}">
                                <div class="p-3">
                                    <div class="mb-2 relative flex mb-4">
                                        <div class="w-1/4">
                                            <button
                                                class="float-left"
                                                on:click="{deleteNote(index, i)}">
                                                <TimesIcon
                                                    color="{note.color}" />
                                            </button>
                                        </div>
                                        <div class="w-2/4">
                                            <input
                                                type="text"
                                                value="{note.title}"
                                                on:change="{noteContentEdit(index, i, 'title')}"
                                                class="inline font-bold text-xl
                                                bg-transparent" />
                                        </div>
                                        <div class="w-1/4 text-right">
                                            <button
                                                on:click="{showChangeColor(note.id)}">
                                                <DropperIcon
                                                    color="{note.color}" />
                                            </button>
                                            {#if changeColor === note.id}
                                                <div
                                                    class="shadow border
                                                    bg-white absolute right-0
                                                    top-0">
                                                    {#each cardColors as color}
                                                        <button
                                                            on:click="{changeNoteColor(index, i, color)}"
                                                            class="p-3 hover:bg-{color}-200
                                                            bg-{color}-100"></button>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>
                                    </div>
                                    <textarea
                                        style="height: {note.contentHeight ? `${note.contentHeight}px` : 'auto'}"
                                        on:mousedown="{detectElementMouseEnlargement(index, i)}"
                                        class="w-full h-full bg-transparent"
                                        rows="5"
                                        on:change="{noteContentEdit(index, i, 'content')}">
                                        {note.content}
                                    </textarea>
                                </div>
                            </li>
                        {/each}
                    </ul>
                </div>
            {/each}
            <div class="m-3 bg-gray-300 w-16 self-stretch flex-no-shrink">
                <button
                    on:click="{addNoteColumn}"
                    class="w-full h-full font-bold text-5xl text-grey">
                    +
                </button>
            </div>
        </section>
    </div>
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
