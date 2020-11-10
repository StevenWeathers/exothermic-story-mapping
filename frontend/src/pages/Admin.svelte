<script>
    import { onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import CreateUser from '../components/CreateUser.svelte'
    import { user } from '../stores.js'
    import { appRoutes } from '../config'
    import { _ } from '../i18n'

    export let xfetch
    export let router
    export let notifications
    export let eventTag

    let appStats = {
        unregisteredUserCount: 0,
        registeredUserCount: 0,
        storyboardCount: 0,
    }
    let users = []
    let showCreateUser = false

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
                eventTag('create_account', 'engagement', 'success')

                getUsers()
                toggleCreateUser()
            })
            .catch(function(error) {
                notifications.danger('Error encountered creating user')
                eventTag('create_account', 'engagement', 'failure')
            })
    }

    xfetch('/api/admin/stats')
        .then(res => res.json())
        .then(function(result) {
            appStats = result
        })
        .catch(function(error) {
            notifications.danger('Error getting application stats')
        })

    function getUsers() {
        xfetch('/api/admin/users')
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

    onMount(() => {
        if (!$user.id) {
            router.route(appRoutes.register)
        }
        if ($user.type !== 'ADMIN') {
            router.route(appRoutes.landing)
        }

        getUsers()
    })
</script>

<PageLayout>
    <div class="text-center px-2 mb-4">
        <h1 class="text-3xl md:text-4xl font-bold">Admin</h1>
    </div>

    <div class="flex justify-center mb-4">
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
                        <th class="w-3/12 px-4 py-2">User Name</th>
                        <th class="w-3/12 px-4 py-2">User Email</th>
                        <th class="w-2/12 px-4 py-2">Verified</th>
                        <th class="w-2/12 px-4 py-2">Type</th>
                        <th class="w-2/12 px-4 py-2">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {#each users as user}
                        <tr>
                            <td class="border px-4 py-2">{user.name}</td>
                            <td class="border px-4 py-2">{user.email}</td>
                            <td class="border px-4 py-2">{user.verified}</td>
                            <td class="border px-4 py-2">{user.type}</td>
                            <td class="border px-4 py-2">
                                {#if user.type !== 'ADMIN'}
                                    <HollowButton
                                        onClick="{promoteUser(user.id)}"
                                        color="blue">
                                        {$_('actions.user.promote')}
                                    </HollowButton>
                                {:else}
                                    <HollowButton
                                        onClick="{demoteUser(user.id)}"
                                        color="blue">
                                        {$_('actions.user.demote')}
                                    </HollowButton>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    </div>

    {#if showCreateUser}
        <CreateUser
            toggleCreate="{toggleCreateUser}"
            handleCreate="{createUser}"
            notifications />
    {/if}
</PageLayout>
