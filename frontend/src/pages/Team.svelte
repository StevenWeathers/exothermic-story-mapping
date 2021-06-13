<script>
    import { onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import AddUser from '../components/AddUser.svelte'
    import RemoveUser from '../components/RemoveUser.svelte'
    import RemoveStoryboard from '../components/RemoveStoryboard.svelte'
    import ChevronRight from '../components/icons/ChevronRight.svelte'
    import CreateStoryboard from '../components/CreateStoryboard.svelte'
    import { user } from '../stores.js'
    import { _ } from '../i18n'
    import { appRoutes } from '../config'

    export let xfetch
    export let router
    export let notifications
    export let eventTag
    export let organizationId
    export let departmentId
    export let teamId

    const storyboardsPageLimit = 1000
    const usersPageLimit = 1000

    let team = {
        id: teamId,
        name: '',
    }
    let organization = {
        id: organizationId,
        name: '',
    }
    let department = {
        id: departmentId,
        name: '',
    }
    let users = []
    let storyboards = []
    let showAddUser = false
    let showRemoveUser = false
    let showRemoveStoryboard = false
    let removeStoryboardId = null
    let removeUserId = null
    let usersPage = 1
    let storyboardsPage = 1

    let organizationRole = ''
    let departmentRole = ''
    let teamRole = ''

    const apiPrefix = '/api'
    $: orgPrefix = departmentId
        ? `${apiPrefix}/organization/${organizationId}/department/${departmentId}`
        : `${apiPrefix}/organization/${organizationId}`
    $: teamPrefix = organizationId
        ? `${orgPrefix}/team/${teamId}`
        : `${apiPrefix}/team/${teamId}`

    function toggleAddUser() {
        showAddUser = !showAddUser
    }

    const toggleRemoveUser = userId => () => {
        showRemoveUser = !showRemoveUser
        removeUserId = userId
    }

    const toggleRemoveStoryboard = storyboardId => () => {
        showRemoveStoryboard = !showRemoveStoryboard
        removeStoryboardId = storyboardId
    }

    function getTeam() {
        xfetch(teamPrefix)
            .then(res => res.json())
            .then(function(result) {
                team = result.team
                teamRole = result.teamRole

                if (departmentId) {
                    department = result.department
                    departmentRole = result.departmentRole
                }
                if (organizationId) {
                    organization = result.organization
                    organizationRole = result.organizationRole
                }

                getStoryboards()
                getUsers()
            })
            .catch(function(error) {
                notifications.danger('Error getting team')
            })
    }

    function getUsers() {
        const usersOffset = (usersPage - 1) * usersPageLimit
        xfetch(`${teamPrefix}/users/${usersPageLimit}/${usersOffset}`)
            .then(res => res.json())
            .then(function(result) {
                users = result
            })
            .catch(function(error) {
                notifications.danger('Error getting team users')
            })
    }

    function getStoryboards() {
        const storyboardsOffset = (storyboardsPage - 1) * storyboardsPageLimit
        xfetch(
            `${teamPrefix}/storyboards/${storyboardsPageLimit}/${storyboardsOffset}`,
        )
            .then(res => res.json())
            .then(function(result) {
                storyboards = result
            })
            .catch(function(error) {
                notifications.danger('Error getting team storyboards')
            })
    }

    function handleUserAdd(email, role) {
        const body = {
            email,
            role,
        }

        xfetch(`${teamPrefix}/users`, { body })
            .then(function() {
                eventTag('team_add_user', 'engagement', 'success')
                toggleAddUser()
                notifications.success('User added successfully.')
                getUsers()
            })
            .catch(function() {
                notifications.danger('Error attempting to add user to team')
                eventTag('team_add_user', 'engagement', 'failure')
            })
    }

    function handleUserRemove() {
        const body = {
            id: removeUserId,
        }

        xfetch(`${teamPrefix}/user`, { body, method: 'DELETE' })
            .then(function() {
                eventTag('team_remove_user', 'engagement', 'success')
                toggleRemoveUser(null)()
                notifications.success('User removed successfully.')
                getUsers()
            })
            .catch(function() {
                notifications.danger(
                    'Error attempting to remove user from team',
                )
                eventTag('team_remove_user', 'engagement', 'failure')
            })
    }

    function handleStoryboardRemove() {
        const body = {
            id: removeStoryboardId,
        }

        xfetch(`${teamPrefix}/storyboard`, { body, method: 'DELETE' })
            .then(function() {
                eventTag('team_remove_storyboard', 'engagement', 'success')
                toggleRemoveStoryboard(null)()
                notifications.success('Storyboard removed successfully.')
                getStoryboards()
            })
            .catch(function() {
                notifications.danger(
                    'Error attempting to remove storyboard from team',
                )
                eventTag('team_remove_storyboard', 'engagement', 'failure')
            })
    }

    onMount(() => {
        if (!$user.id || $user.type === 'GUEST') {
            router.route(appRoutes.login)
        }

        getTeam()
    })

    $: isAdmin =
        organizationRole === 'ADMIN' ||
        departmentRole === 'ADMIN' ||
        teamRole === 'ADMIN'
    $: isTeamMember =
        organizationRole === 'ADMIN' ||
        departmentRole === 'ADMIN' ||
        teamRole !== ''
</script>

<svelte:head>
    <title>Team {team.name} | Exothermic</title>
</svelte:head>

<PageLayout>
    <h1 class="text-3xl font-bold">Team: {team.name}</h1>
    {#if organizationId}
        <div class="font-bold">
            Organization
            <ChevronRight class="inline-block" />
            <a
                class="text-blue-500 hover:text-blue-800"
                href="{appRoutes.organization}/{organization.id}">
                {organization.name}
            </a>
            {#if departmentId}
                &nbsp;
                <ChevronRight class="inline-block" />
                Department
                <ChevronRight class="inline-block" />
                <a
                    class="text-blue-500 hover:text-blue-800"
                    href="{appRoutes.organization}/{organization.id}/department/{department.id}">
                    {department.name}
                </a>
            {/if}
        </div>
    {/if}

    <div class="w-full mt-4">
        <div class="p-4 md:p-6 bg-white shadow-lg rounded flex">
            <div class="w-full md:w-1/2 lg:w-3/5 md:pr-4">
                <div class="flex w-full">
                    <h2 class="text-2xl md:text-3xl font-bold mb-4">
                        Storyboards
                    </h2>
                </div>

                <table class="table-fixed w-full">
                    <thead>
                        <tr>
                            <th class="w-2/6 px-4 py-2">Name</th>
                            <th class="w-1/6 px-4 py-2"></th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each storyboards as storyboard}
                            <tr>
                                <td class="border px-4 py-2">
                                    {storyboard.name}
                                </td>
                                <td class="border px-4 py-2 text-right">
                                    {#if isAdmin}
                                        <HollowButton
                                            onClick="{toggleRemoveStoryboard(storyboard.id)}"
                                            color="red">
                                            Remove
                                        </HollowButton>
                                    {/if}
                                    <HollowButton
                                        href="{appRoutes.storyboard}/{storyboard.id}">
                                        Join
                                    </HollowButton>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>

            <div class="w-full md:w-1/2 lg:w-2/5 md:pl-2 xl:pl-4">
                {#if isTeamMember}
                    <h2 class="mb-4 text-2xl font-bold leading-tight">
                        Create Storyboard
                    </h2>
                    <CreateStoryboard
                        apiPrefix="{teamPrefix}"
                        {notifications}
                        {router}
                        {eventTag}
                        {xfetch} />
                {/if}
            </div>
        </div>
    </div>

    <div class="w-full mt-4">
        <div class="p-4 md:p-6 bg-white shadow-lg rounded">
            <div class="flex w-full">
                <div class="w-4/5">
                    <h2 class="text-2xl md:text-3xl font-bold mb-4">Users</h2>
                </div>
                <div class="w-1/5">
                    <div class="text-right">
                        {#if isAdmin}
                            <HollowButton onClick="{toggleAddUser}">
                                Add User
                            </HollowButton>
                        {/if}
                    </div>
                </div>
            </div>

            <table class="table-fixed w-full">
                <thead>
                    <tr>
                        <th class="w-2/6 px-4 py-2">Name</th>
                        <th class="w-2/6 px-4 py-2">Email</th>
                        <th class="w-1/6 px-4 py-2">Role</th>
                        <th class="w-1/6 px-4 py-2"></th>
                    </tr>
                </thead>
                <tbody>
                    {#each users as usr}
                        <tr>
                            <td class="border px-4 py-2">{usr.name}</td>
                            <td class="border px-4 py-2">{usr.email}</td>
                            <td class="border px-4 py-2">{usr.role}</td>
                            <td class="border px-4 py-2 text-right">
                                {#if isAdmin}
                                    <HollowButton
                                        onClick="{toggleRemoveUser(usr.id)}"
                                        color="red">
                                        Remove
                                    </HollowButton>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    </div>

    {#if showAddUser}
        <AddUser toggleAdd="{toggleAddUser}" handleAdd="{handleUserAdd}" />
    {/if}

    {#if showRemoveUser}
        <RemoveUser
            toggleRemove="{toggleRemoveUser(null)}"
            handleRemove="{handleUserRemove}" />
    {/if}

    {#if showRemoveStoryboard}
        <RemoveStoryboard
            toggleRemove="{toggleRemoveStoryboard(null)}"
            handleRemove="{handleStoryboardRemove}" />
    {/if}
</PageLayout>
