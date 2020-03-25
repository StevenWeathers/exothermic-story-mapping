<script>
    import { onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import CreateStoryboard from '../components/CreateStoryboard.svelte'
    import DownCarrotIcon from '../components/icons/DownCarrotIcon.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import { user } from '../stores.js'

    export let notifications
    export let router
    let storyboards = []
    let storyboardName = ''

    fetch('/api/storyboards', {
        method: 'GET',
        credentials: 'same-origin',
    })
        .then(function(response) {
            if (!response.ok) {
                throw Error(response.statusText)
            }
            return response
        })
        .then(function(response) {
            return response.json()
        })
        .then(function(bs) {
            storyboards = bs
        })
        .catch(function(error) {
            notifications.danger('Error finding your storyboards')
        })

    function createStoryboard(e) {
        e.preventDefault()
        const data = {
            storyboardName,
            ownerId: $user.id,
        }

        fetch('/api/storyboard', {
            method: 'POST',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
            .then(function(response) {
                if (!response.ok) {
                    throw Error(response.statusText)
                }
                return response
            })
            .then(function(response) {
                return response.json()
            })
            .then(function(storyboard) {
                router.route(`/storyboard/${storyboard.id}`)
            })
            .catch(function(error) {
                notifications.danger('Error encountered creating storyboard')
            })
    }

    onMount(() => {
        if (!$user.id) {
            router.route('/register')
        }
    })
</script>

<PageLayout>
    <h1 class="mb-4 text-3xl font-bold">My Storyboards</h1>

    <div class="flex flex-wrap">
        <div class="mb-4 md:mb-6 w-full md:w-1/2 lg:w-3/5 md:pr-4">
            {#each storyboards as storyboard}
                <div class="bg-white shadow-lg rounded mb-2">
                    <div
                        class="flex flex-wrap items-center p-4 border-gray-400
                        border-b">
                        <div
                            class="w-full md:w-1/2 mb-4 md:mb-0 font-semibold
                            md:text-xl leading-tight">
                            {storyboard.name}
                            <div class="font-semibold md:text-sm text-gray-600">
                                {#if $user.id === storyboard.ownerId}Owner{/if}
                            </div>
                        </div>
                        <div class="w-full md:w-1/2 md:mb-0 md:text-right">
                            <HollowButton href="/storyboard/{storyboard.id}">
                                Join Storyboard
                            </HollowButton>
                        </div>
                    </div>
                </div>
            {/each}
        </div>

        <div class="w-full md:w-1/2 lg:w-2/5 pl-4">
            <div class="p-6 bg-white shadow-lg rounded">
                <h2 class="mb-4 text-2xl font-bold leading-tight">
                    Create a Storyboard
                </h2>
                <CreateStoryboard {notifications} {router} />
            </div>
        </div>
    </div>
</PageLayout>
