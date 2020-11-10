<script>
    import { onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import UserAvatar from '../components/UserAvatar.svelte'
    import DownCarrotIcon from '../components/icons/DownCarrotIcon.svelte'
    import { user } from '../stores.js'
    import { validateName, validatePasswords } from '../validationUtils.js'
    import { appRoutes } from '../config'
    import CreateApiKey from '../components/CreateApiKey.svelte'
    import { _ } from '../i18n'

    export let xfetch
    export let router
    export let eventTag
    export let notifications

    let userProfile = {}
    let apiKeys = []
    let showApiKeyCreate = false

    let updatePassword = false
    let userPassword1 = ''
    let userPassword2 = ''

    const { APIEnabled, AvatarService, AuthMethod } = appConfig
    const configurableAvatarServices = [
        'dicebear',
        'gravatar',
        'robohash',
        'govatar',
    ]
    const isAvatarConfigurable = configurableAvatarServices.includes(
        AvatarService,
    )
    const avatarOptions = {
        dicebear: [
            'male',
            'female',
            'human',
            'identicon',
            'bottts',
            'avataaars',
            'jdenticon',
            'gridy',
            'code',
        ],
        gravatar: [
            'mp',
            'identicon',
            'monsterid',
            'wavatar',
            'retro',
            'robohash',
        ],
        robohash: ['set1', 'set2', 'set3', 'set4'],
        govatar: ['male', 'female'],
    }

    let avatars = isAvatarConfigurable ? avatarOptions[AvatarService] : []

    function toggleUpdatePassword() {
        updatePassword = !updatePassword
        eventTag(
            'update_password_toggle',
            'engagement',
            `update: ${updatePassword}`,
        )
    }

    xfetch(`/api/user/${$user.id}`)
        .then(res => res.json())
        .then(function(wp) {
            userProfile = wp
        })
        .catch(function(error) {
            notifications.danger('Error getting your profile')
            eventTag('fetch_profile', 'engagement', 'failure')
        })

    function updateUserProfile(e) {
        e.preventDefault()
        const body = {
            userName: userProfile.name,
            UserAvatar: userProfile.avatar,
        }
        const validName = validateName(body.userName)

        let noFormErrors = true

        if (!validName.valid) {
            noFormErrors = false
            notifications.danger(validName.error, 1500)
        }

        if (noFormErrors) {
            xfetch(`/api/user/${$user.id}`, { body })
                .then(function(updatedUser) {
                    user.update({
                        id: userProfile.id,
                        name: userProfile.name,
                        email: userProfile.email,
                        type: userProfile.type,
                        avatar: userProfile.avatar,
                    })

                    notifications.success('Profile updated.', 1500)
                    eventTag('update_profile', 'engagement', 'success')
                })
                .catch(function(error) {
                    notifications.danger(
                        'Error encountered updating your profile',
                    )
                    eventTag('update_profile', 'engagement', 'failure')
                })
        }
    }

    function updateUserPassword(e) {
        e.preventDefault()
        const body = {
            userPassword1,
            userPassword2,
        }
        const validPasswords = validatePasswords(userPassword1, userPassword2)

        let noFormErrors = true

        if (!validPasswords.valid) {
            noFormErrors = false
            notifications.danger(validPasswords.error, 1500)
        }

        if (noFormErrors) {
            xfetch('/api/auth/update-password', { body })
                .then(function() {
                    notifications.success('Password updated.', 1500)
                    updatePassword = false
                    eventTag('update_password', 'engagement', 'success')
                })
                .catch(function(error) {
                    notifications.danger(
                        'Error encountered attempting to update password',
                    )
                    eventTag('update_password', 'engagement', 'failure')
                })
        }
    }

    function getApiKeys() {
        xfetch(`/api/user/${$user.id}/apikeys`)
            .then(res => res.json())
            .then(function(apks) {
                apiKeys = apks
            })
            .catch(function(error) {
                notifications.danger(
                    $_('pages.userProfile.apiKeys.errorRetreiving'),
                )
                eventTag('fetch_profile_apikeys', 'engagement', 'failure')
            })
    }
    getApiKeys()

    function deleteApiKey(apk) {
        return function() {
            xfetch(`/api/user/${$user.id}/apikey/${apk}`, {
                method: 'DELETE',
            })
                .then(res => res.json())
                .then(function(apks) {
                    notifications.success(
                        $_('pages.userProfile.apiKeys.deleteSuccess'),
                    )
                    apiKeys = apks
                })
                .catch(function(error) {
                    notifications.danger(
                        $_('pages.userProfile.apiKeys.deleteFailed'),
                    )
                })
        }
    }

    function toggleApiKeyActiveStatus(apk, active) {
        return function() {
            const body = {
                active: !active,
            }

            xfetch(`/api/user/${$user.id}/apikey/${apk}`, {
                body,
                method: 'PUT',
            })
                .then(res => res.json())
                .then(function(apks) {
                    notifications.success(
                        $_('pages.userProfile.apiKeys.updateSuccess'),
                    )
                    apiKeys = apks
                })
                .catch(function(error) {
                    notifications.danger(
                        $_('pages.userProfile.apiKeys.updateFailed'),
                    )
                })
        }
    }

    function toggleCreateApiKey() {
        showApiKeyCreate = !showApiKeyCreate
    }

    onMount(() => {
        if (!$user.id) {
            router.route(appRoutes.register)
        }
    })

    $: updateDisabled = userProfile.name === ''
    $: updatePasswordDisabled =
        userPassword1 === '' || userPassword2 === '' || AuthMethod === 'ldap'
</script>

<PageLayout>
    <div class="flex justify-center flex-wrap">
        <div class="w-full md:w-1/2 lg:w-1/3">
            {#if !updatePassword}
                <form
                    on:submit="{updateUserProfile}"
                    class="bg-white shadow-lg rounded p-4 md:p-6 mb-4"
                    name="updateProfile">
                    <h2
                        class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                        md:leading-tight">
                        Your Profile
                    </h2>

                    <div class="mb-4">
                        <label
                            class="block text-gray-700 text-sm font-bold mb-2"
                            for="yourName">
                            Name
                        </label>
                        <input
                            bind:value="{userProfile.name}"
                            placeholder="Enter your name"
                            class="bg-gray-200 border-gray-200 border-2
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-orange-500"
                            id="yourName"
                            name="yourName"
                            type="text"
                            required />
                    </div>

                    <div class="mb-4">
                        <label
                            class="block text-gray-700 text-sm font-bold mb-2"
                            for="yourEmail">
                            Email
                            {#if userProfile.verified}
                                <span
                                    class="font-bold text-green-600
                                    border-green-500 border py-1 px-2 rounded
                                    ml-1">
                                    Verified
                                </span>
                            {/if}
                        </label>
                        <input
                            bind:value="{userProfile.email}"
                            class="bg-gray-200 border-gray-200 border-2
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            cursor-not-allowed"
                            id="yourEmail"
                            name="yourEmail"
                            type="email"
                            disabled />
                    </div>

                    {#if isAvatarConfigurable}
                        <div class="mb-4">
                            <label
                                class="block text-gray-700 text-sm font-bold
                                mb-2"
                                for="yourAvatar">
                                Avatar
                            </label>
                            <div class="flex">
                                <div class="md:w-2/3 lg:w-3/4">
                                    <div class="relative">
                                        <select
                                            bind:value="{userProfile.avatar}"
                                            class="block appearance-none w-full
                                            border-2 border-gray-400
                                            text-gray-700 py-3 px-4 pr-8 rounded
                                            leading-tight focus:outline-none
                                            focus:border-purple-500"
                                            id="yourAvatar"
                                            name="yourAvatar">
                                            {#each avatars as item}
                                                <option value="{item}">
                                                    {item}
                                                </option>
                                            {/each}
                                        </select>
                                        <div
                                            class="pointer-events-none absolute
                                            inset-y-0 right-0 flex items-center
                                            px-2 text-gray-700">
                                            <DownCarrotIcon />
                                        </div>
                                    </div>
                                </div>
                                <div class="md:w-1/3 lg:w-1/4 ml-1">
                                    <span class="float-right">
                                        <UserAvatar
                                            userId="{userProfile.id}"
                                            avatar="{userProfile.avatar}"
                                            {AvatarService}
                                            width="40" />
                                    </span>
                                </div>
                            </div>
                        </div>
                    {/if}

                    <div>
                        <div class="text-right">
                            <button
                                type="button"
                                class="inline-block align-baseline font-bold
                                text-sm text-blue-500 hover:text-blue-800 mr-4"
                                on:click="{toggleUpdatePassword}">
                                Update Password
                            </button>
                            <SolidButton
                                type="submit"
                                disabled="{updateDisabled}">
                                Update Profile
                            </SolidButton>
                        </div>
                    </div>
                </form>
            {/if}

            {#if updatePassword}
                <form
                    on:submit="{updateUserPassword}"
                    class="bg-white shadow-lg rounded p-6 mb-4"
                    name="updateUserPassword">
                    <div
                        class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                        md:leading-tight text-center">
                        Update Password
                    </div>

                    <div class="mb-4">
                        <label
                            class="block text-gray-700 text-sm font-bold mb-2"
                            for="yourPassword1">
                            Password
                        </label>
                        <input
                            bind:value="{userPassword1}"
                            placeholder="Enter a password"
                            class="bg-gray-200 border-gray-200 border-2
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-purple-500"
                            id="yourPassword1"
                            name="yourPassword1"
                            type="password"
                            required />
                    </div>

                    <div class="mb-4">
                        <label
                            class="block text-gray-700 text-sm font-bold mb-2"
                            for="yourPassword2">
                            Confirm Password
                        </label>
                        <input
                            bind:value="{userPassword2}"
                            placeholder="Confirm your password"
                            class="bg-gray-200 border-gray-200 border-2
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-purple-500"
                            id="yourPassword2"
                            name="yourPassword2"
                            type="password"
                            required />
                    </div>

                    <div class="text-right">
                        <button
                            type="button"
                            class="inline-block align-baseline font-bold text-sm
                            text-blue-500 hover:text-blue-800 mr-4"
                            on:click="{toggleUpdatePassword}">
                            Cancel
                        </button>
                        <SolidButton
                            type="submit"
                            disabled="{updatePasswordDisabled}">
                            Update
                        </SolidButton>
                    </div>
                </form>
            {/if}
        </div>

        {#if APIEnabled}
            <div class="w-full">
                <div class="bg-white shadow-lg rounded p-4 md:p-6 mb-4">
                    <div class="flex w-full">
                        <div class="w-4/5">
                            <h2
                                class="text-2xl md:text-3xl font-bold
                                text-center mb-4">
                                {$_('pages.userProfile.apiKeys.title')}
                            </h2>
                        </div>
                        <div class="w-1/5">
                            <div class="text-right">
                                <HollowButton onClick="{toggleCreateApiKey}">
                                    {$_('pages.userProfile.apiKeys.createButton')}
                                </HollowButton>
                            </div>
                        </div>
                    </div>

                    <table class="table-fixed w-full">
                        <thead>
                            <tr>
                                <th class="w-2/12 px-4 py-2">
                                    {$_('pages.userProfile.apiKeys.name')}
                                </th>
                                <th class="w-2/12 px-4 py-2">
                                    {$_('pages.userProfile.apiKeys.prefix')}
                                </th>
                                <th class="w-2/12 px-4 py-2">
                                    {$_('pages.userProfile.apiKeys.active')}
                                </th>
                                <th class="w-3/12 px-4 py-2">
                                    {$_('pages.userProfile.apiKeys.updated')}
                                </th>
                                <th class="w-3/12 px-4 py-2">
                                    {$_('pages.userProfile.apiKeys.actions')}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each apiKeys as apk}
                                <tr>
                                    <td class="border px-4 py-2">{apk.name}</td>
                                    <td class="border px-4 py-2">
                                        {apk.prefix}
                                    </td>
                                    <td class="border px-4 py-2">
                                        {apk.active}
                                    </td>
                                    <td class="border px-4 py-2">
                                        {new Date(apk.updatedDate).toLocaleString()}
                                    </td>
                                    <td class="border px-4 py-2">
                                        <HollowButton
                                            onClick="{toggleApiKeyActiveStatus(apk.id, apk.active)}">
                                            {#if !apk.active}
                                                {$_('pages.userProfile.apiKeys.activateButton')}
                                            {:else}
                                                {$_('pages.userProfile.apiKeys.deactivateButton')}
                                            {/if}
                                        </HollowButton>
                                        <HollowButton
                                            color="red"
                                            onClick="{deleteApiKey(apk.id)}">
                                            {$_('pages.userProfile.apiKeys.deleteButton')}
                                        </HollowButton>
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            </div>

            {#if showApiKeyCreate}
                <CreateApiKey
                    {toggleCreateApiKey}
                    handleApiKeyCreate="{getApiKeys}"
                    {notifications}
                    {xfetch}
                    {eventTag} />
            {/if}
        {/if}
    </div>
</PageLayout>
