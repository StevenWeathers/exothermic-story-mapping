<script>
    import { onMount } from 'svelte'

    import AdminPageLayout from '../../components/AdminPageLayout.svelte'
    import HollowButton from '../../components/HollowButton.svelte'
    import CreateUser from '../../components/CreateUser.svelte'
    import Pagination from '../../components/Pagination.svelte'
    import CountryFlag from '../../components/CountryFlag.svelte'
    import CheckIcon from '../../components/icons/CheckIcon.svelte'
    import { user } from '../../stores.js'
    import { _ } from '../../i18n'
    import { appRoutes } from '../../config'

    export let xfetch
    export let router
    export let notifications
    export let eventTag

    const usersPageLimit = 100

    let appStats = {
        unregisteredUserCount: 0,
        registeredUserCount: 0,
        battleCount: 0,
        planCount: 0,
        organizationCount: 0,
        departmentCount: 0,
        teamCount: 0,
    }
    let users = []
    let showCreateUser = false
    let usersPage = 1

    function toggleCreateUser() {
        showCreateUser = !showCreateUser
    }

    function createUser(userName, userEmail, userPassword1, userPassword2) {
        const body = {
            userName,
            userEmail,
            userPassword1,
            userPassword2,
        }

        xfetch('/api/admin/user', { body })
            .then(function() {
                eventTag('admin_create_user', 'engagement', 'success')

                getUsers()
                toggleCreateUser()
            })
            .catch(function(error) {
                notifications.danger('Error encountered creating user')
                eventTag('admin_create_user', 'engagement', 'failure')
            })
    }

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

    function getUsers() {
        const usersOffset = (usersPage - 1) * usersPageLimit
        xfetch(`/api/admin/users/${usersPageLimit}/${usersOffset}`)
            .then(res => res.json())
            .then(function(result) {
                users = result
            })
            .catch(function(error) {
                notifications.danger('Error getting users')
            })
    }

    function promoteUser(userId) {
        return function() {
            const body = {
                userId,
            }

            xfetch('/api/admin/promote', { body })
                .then(function() {
                    eventTag('admin_promote_user', 'engagement', 'success')

                    getUsers()
                })
                .catch(function(error) {
                    notifications.danger('Error encountered promoting user')
                    eventTag('admin_promote_user', 'engagement', 'failure')
                })
        }
    }

    function demoteUser(userId) {
        return function() {
            const body = {
                userId,
            }

            xfetch('/api/admin/demote', { body })
                .then(function() {
                    eventTag('admin_demote_user', 'engagement', 'success')

                    getUsers()
                })
                .catch(function(error) {
                    notifications.danger('Error encountered demoting user')
                    eventTag('admin_demote_user', 'engagement', 'failure')
                })
        }
    }

    const changePage = evt => {
        usersPage = evt.detail
        getUsers()
    }

    onMount(() => {
        if (!$user.id) {
            router.route(appRoutes.login)
        }
        if ($user.type !== 'ADMIN') {
            router.route(appRoutes.landing)
        }

        getAppStats()
        getUsers()
    })
</script>

<svelte:head>
    <title>Users Admin | Exothermic</title>
</svelte:head>

<AdminPageLayout activePage="users">
    <div class="text-center px-2 mb-4">
        <h1 class="text-3xl md:text-4xl font-bold">Users</h1>
    </div>

    <div class="w-full">
        <div class="p-4 md:p-6 bg-white shadow-lg rounded">
            <div class="flex w-full">
                <div class="w-4/5">
                    <h2 class="text-2xl md:text-3xl font-bold text-center mb-4">
                        Registered Users
                    </h2>
                </div>
                <div class="w-1/5">
                    <div class="text-right">
                        <HollowButton onClick="{toggleCreateUser}">
                            Create User
                        </HollowButton>
                    </div>
                </div>
            </div>

            <table class="table-fixed w-full">
                <thead>
                    <tr>
                        <th class="w-3/12 p-2">Name</th>
                        <th class="w-3/12 p-2">Email</th>
                        <th class="w-3/12 p-2">Company</th>
                        <th class="w-2/12 p-2">Type</th>
                        <th class="w-1/12 p-2"></th>
                    </tr>
                </thead>
                <tbody>
                    {#each users as user}
                        <tr>
                            <td class="border p-2">
                                {user.name}
                                {#if user.country}
                                    &nbsp;
                                    <CountryFlag
                                        country="{user.country}"
                                        size="{16}"
                                        additionalClass="inline-block" />
                                {/if}
                            </td>
                            <td class="border p-2">
                                {user.email}
                                {#if user.verified}
                                    &nbsp;
                                    <span
                                        class="text-green-600"
                                        title="Verified">
                                        <CheckIcon />
                                    </span>
                                {/if}
                            </td>
                            <td class="border p-2">
                                <div>{user.company}</div>
                                {#if user.jobTitle}
                                    <div class="text-gray-700 text-sm">
                                        Job Title: {user.jobTitle}
                                    </div>
                                {/if}
                            </td>
                            <td class="border p-2">{user.type}</td>
                            <td class="border p-2">
                                {#if user.type !== 'ADMIN'}
                                    <HollowButton
                                        onClick="{promoteUser(user.id)}"
                                        color="blue">
                                        Promote
                                    </HollowButton>
                                {:else}
                                    <HollowButton
                                        onClick="{demoteUser(user.id)}"
                                        color="blue">
                                        Demote
                                    </HollowButton>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>

            {#if appStats.registeredUserCount > usersPageLimit}
                <div class="pt-6 flex justify-center">
                    <Pagination
                        bind:current="{usersPage}"
                        num_items="{appStats.registeredUserCount}"
                        per_page="{usersPageLimit}"
                        on:navigate="{changePage}" />
                </div>
            {/if}
        </div>
    </div>

    {#if showCreateUser}
        <CreateUser
            toggleCreate="{toggleCreateUser}"
            handleCreate="{createUser}"
            notifications />
    {/if}
</AdminPageLayout>
