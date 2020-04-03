<script>
    import { onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import { user } from '../stores.js'

    export let router
    export let notifications

    let appStats = {
        unregisteredUserCount: 0,
        registeredUserCount: 0,
        storyboardCount: 0,
    }

    fetch('/api/admin/stats', {
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
        .then(function(result) {
            appStats = result
        })
        .catch(function(error) {
            notifications.danger('Error getting application stats')
        })

    onMount(() => {
        if (!$user.id) {
            router.route('/enlist')
        }
        if ($user.type !== 'ADMIN') {
            router.route('/')
        }
    })
</script>

<PageLayout>
    <div class="text-center px-2 mb-4">
        <h1 class="text-3xl md:text-4xl font-bold">Admin</h1>
    </div>
    <div class="flex justify-center">
        <div class="w-full">
            <div
                class="flex flex-wrap items-center text-center pt-2 pb-2 md:pt-4
                md:pb-4 bg-white shadow-lg rounded text-xl">
                <div class="w-1/3">
                    <div class="mb-2 font-bold">Unregistered Users</div>
                    {appStats.unregisteredUserCount}
                </div>
                <div class="w-1/3">
                    <div class="mb-2 font-bold">Registered Users</div>
                    {appStats.registeredUserCount}
                </div>
                <div class="w-1/3">
                    <div class="mb-2 font-bold">Storyboards</div>
                    {appStats.storyboardCount}
                </div>
            </div>
        </div>
    </div>
</PageLayout>
