<script>
    import Sockette from 'sockette'
    import { onMount, onDestroy } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import UserCard from '../components/UserCard.svelte'
    import InviteUser from '../components/InviteUser.svelte'
    import HollowButton from '../components/HollowButton.svelte'

    import { user } from '../stores.js'

    export let storyboardId
    export let notifications
    export let router

    const hostname = window.location.origin
    const socketExtension = window.location.protocol === 'https:' ? 'wss' : 'ws'

    let socketError = false
    let socketReconnecting = false
    let storyboard = {}

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

    onMount(() => {
        if (!$user.id) {
            router.route(`/register/${storyboardId}`)
        }
    })
</script>

<svelte:head>
    <title>Storyboard {storyboard.name} | Exothermic</title>
</svelte:head>

<PageLayout>
    {#if storyboard.name && !socketReconnecting && !socketError}
        <div class="mb-6 flex flex-wrap">
            <div class="w-full text-center text-left">
                <h1 class="text-3xl font-bold leading-tight">
                    {storyboard.name}
                </h1>
            </div>
        </div>

        <div class="flex flex-wrap mb-4 -mx-4">
            <div class="w-full lg:w-3/4 px-4"></div>

            <div class="w-full lg:w-1/4 px-4">
                <div class="bg-white shadow-lg mb-4 rounded">
                    <div class="bg-blue-500 p-4 rounded-t">
                        <h3 class="text-2xl text-white leading-tight font-bold">
                            Users
                        </h3>
                    </div>

                    {#each storyboard.users as usr (usr.id)}
                        {#if usr.active}
                            <UserCard user="{usr}" {sendSocketEvent} />
                        {/if}
                    {/each}
                </div>

                <div class="bg-white shadow-lg p-4 mb-4 rounded">
                    <InviteUser {hostname} storyboardId="{storyboard.id}" />
                    {#if storyboard.ownerId === $user.id}
                        <div class="mt-4 text-right">
                            <HollowButton
                                color="red"
                                onClick="{concedeStoryboard}">
                                Delete Storyboard
                            </HollowButton>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    {:else if socketReconnecting}
        <div class="flex items-center">
            <div class="flex-1 text-center">
                <h1 class="text-5xl text-orange-500 leading-tight font-bold">
                    Ooops, reloading Storyboard...
                </h1>
            </div>
        </div>
    {:else if socketError}
        <div class="flex items-center">
            <div class="flex-1 text-center">
                <h1 class="text-5xl text-red-500 leading-tight font-bold">
                    Error joining storyboard, refresh and try again.
                </h1>
            </div>
        </div>
    {:else}
        <div class="flex items-center">
            <div class="flex-1 text-center">
                <h1 class="text-5xl text-green-500 leading-tight font-bold">
                    Loading Storyboard...
                </h1>
            </div>
        </div>
    {/if}
</PageLayout>
