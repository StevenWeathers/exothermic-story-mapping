<script>
    import { onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import { user } from '../stores.js'

    export let router
    export let notifications

    const nameMin = 1
    const nameMax = 64

    let userProfile = {}

    fetch(`/api/user/${$user.id}`, {
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
        .then(function(wp) {
            userProfile = wp
        })
        .catch(function(error) {
            notifications.danger('Error getting your profile')
        })

    function updateUserProfile(e) {
        e.preventDefault()
        const body = {
            userName: userProfile.name,
        }

        let noFormErrors = true

        if (body.userName.length < nameMin || body.userName.length > nameMax) {
            noFormErrors = false
            notifications.danger(
                `Name must be between ${nameMin} and ${nameMax} characters.`,
            )
        }

        if (noFormErrors) {
            fetch(`/api/user/${$user.id}`, {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(body),
            })
                .then(function(response) {
                    if (!response.ok) {
                        throw Error(response.statusText)
                    }
                    return response
                })
                .then(function(updatedUser) {
                    user.update({
                        id: userProfile.id,
                        name: userProfile.name,
                        email: userProfile.email,
                        type: userProfile.type,
                    })

                    notifications.success('Profile updated.', 1500)
                })
                .catch(function(error) {
                    notifications.danger(
                        'Error encountered updating your profile',
                    )
                })
        }
    }

    onMount(() => {
        if (!$user.id) {
            router.route('/register')
        }
    })

    $: updateDisabled = userProfile.name === ''
</script>

<PageLayout>
    <div class="flex justify-center">
        <div class="w-full md:w-1/2 lg:w-1/3">
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
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
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
                    </label>
                    <input
                        bind:value="{userProfile.email}"
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none cursor-not-allowed"
                        id="yourEmail"
                        name="yourEmail"
                        type="email"
                        disabled />
                </div>

                <div>
                    <div class="text-right">
                        <SolidButton type="submit" disabled="{updateDisabled}">
                            Update Profile
                        </SolidButton>
                    </div>
                </div>
            </form>
        </div>
    </div>
</PageLayout>
