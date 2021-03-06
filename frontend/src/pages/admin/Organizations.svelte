<script>
    import { onMount } from 'svelte'

    import AdminPageLayout from '../../components/AdminPageLayout.svelte'
    import Pagination from '../../components/Pagination.svelte'
    import { user } from '../../stores.js'
    import { _ } from '../../i18n'
    import { appRoutes } from '../../config'

    export let xfetch
    export let router
    export let notifications

    const organizationsPageLimit = 100

    let appStats = {
        unregisteredUserCount: 0,
        registeredUserCount: 0,
        storyboardCount: 0,
        organizationCount: 0,
        departmentCount: 0,
        teamCount: 0,
    }
    let organizations = []
    let organizationsPage = 1

    function getAppStats() {
        xfetch('/api/admin/stats')
            .then(res => res.json())
            .then(function(result) {
                appStats = result
            })
            .catch(function(error) {
                notifications.danger('Error getting application stats')
            })
    }

    function getOrganizations() {
        const organizationsOffset =
            (organizationsPage - 1) * organizationsPageLimit
        xfetch(
            `/api/admin/organizations/${organizationsPageLimit}/${organizationsOffset}`,
        )
            .then(res => res.json())
            .then(function(result) {
                organizations = result
            })
            .catch(function(error) {
                notifications.danger('Error getting organizations')
            })
    }

    const changePage = evt => {
        organizationsPage = evt.detail
        getOrganizations()
    }

    onMount(() => {
        if (!$user.id) {
            router.route(appRoutes.login)
        }
        if ($user.type !== 'ADMIN') {
            router.route(appRoutes.landing)
        }

        getAppStats()
        getOrganizations()
    })
</script>

<svelte:head>
    <title>Organizations Admin | Exothermic</title>
</svelte:head>

<AdminPageLayout activePage="organizations">
    <div class="text-center px-2 mb-4">
        <h1 class="text-3xl md:text-4xl font-bold">Organizations</h1>
    </div>

    <div class="w-full">
        <div class="p-4 md:p-6 bg-white shadow-lg rounded">
            <table class="table-fixed w-full">
                <thead>
                    <tr>
                        <th class="w-2/6 px-4 py-2">Name</th>
                        <th class="w-1/6 px-4 py-2"></th>
                    </tr>
                </thead>
                <tbody>
                    {#each organizations as org}
                        <tr>
                            <td class="border px-4 py-2">{org.name}</td>
                            <td class="border px-4 py-2"></td>
                        </tr>
                    {/each}
                </tbody>
            </table>

            {#if appStats.organizationCount > organizationsPageLimit}
                <div class="pt-6 flex justify-center">
                    <Pagination
                        bind:current="{organizationsPage}"
                        num_items="{appStats.organizationCount}"
                        per_page="{organizationsPageLimit}"
                        on:navigate="{changePage}" />
                </div>
            {/if}
        </div>
    </div>
</AdminPageLayout>
