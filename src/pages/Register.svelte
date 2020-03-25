<script>
    import PageLayout from '../components/PageLayout.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import { user } from '../stores.js'

    export let router
    export let notifications
    export let storyboardId

    const nameMin = 1
    const nameMax = 64
    const passMin = 6
    const passMax = 72
    const emailMax = 320

    let userName = ''
    let userEmail = ''
    let userPassword1 = ''
    let userPassword2 = ''

    $: targetPage = storyboardId
        ? `/storyboard/${storyboardId}`
        : '/storyboards'

    function createUserGuest(e) {
        e.preventDefault()
        const body = {
            userName,
        }

        let noFormErrors = true

        if (userName.length < nameMin || userName.length > nameMax) {
            noFormErrors = false
            notifications.danger(
                `Name must be between ${nameMin} and ${nameMax} characters.`,
            )
        }

        if (noFormErrors) {
            fetch('/api/user', {
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
                .then(function(response) {
                    return response.json()
                })
                .then(function(newUser) {
                    user.create({
                        id: newUser.id,
                        name: newUser.name,
                        type: newUser.type,
                    })

                    router.route(targetPage, true)
                })
                .catch(function(error) {
                    notifications.danger(
                        'Error encountered registering user as guest',
                    )
                })
        }
    }

    function createUserRegistered(e) {
        e.preventDefault()
        const body = {
            userName,
            userEmail,
            userPassword1,
            userPassword2,
        }

        let noFormErrors = true

        if (userName.length < nameMin || userName.length > nameMax) {
            noFormErrors = false
            notifications.danger(
                `Name must be between ${nameMin} and ${nameMax} characters.`,
            )
        }

        if (userPassword1.length < passMin || userPassword1.length > passMax) {
            noFormErrors = false
            notifications.danger(
                `Password must be between ${passMin} and ${passMax} characters.`,
            )
        }

        if (userPassword1 !== userPassword2) {
            noFormErrors = false
            notifications.danger(`Password and Confirm Password do not match.`)
        }

        if (noFormErrors) {
            fetch('/api/register', {
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
                .then(function(response) {
                    return response.json()
                })
                .then(function(newUser) {
                    user.create({
                        id: newUser.id,
                        name: newUser.name,
                        email: newUser.email,
                        type: newUser.type,
                    })

                    router.route(targetPage, true)
                })
                .catch(function(error) {
                    notifications.danger('Error encountered creating user')
                })
        }
    }

    $: registerDisabled = userName === ''
    $: createDisabled =
        userName === '' ||
        userEmail === '' ||
        userPassword1 === '' ||
        userPassword2 === ''
</script>

<PageLayout>
    <div class="text-center px-2 mb-4">
        <h1 class="text-3xl md:text-4xl font-bold">Enlist to Storyboard</h1>
    </div>
    <div class="flex flex-wrap">
        <div class="w-full md:w-1/2 px-4">
            <form
                on:submit="{createUserGuest}"
                class="bg-white shadow-lg rounded p-4 md:p-6 mb-4"
                name="registerGuest">
                <h2
                    class="font-bold text-xl md:text-2xl b-4 mb-2 md:mb-6
                    md:leading-tight">
                    Register as Guest
                </h2>

                <div class="mb-6">
                    <label
                        class="block text-gray-700 text-sm font-bold mb-2"
                        for="yourName1">
                        Name
                    </label>
                    <input
                        bind:value="{userName}"
                        placeholder="Enter your name"
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
                        id="yourName1"
                        name="yourName1"
                        required />
                </div>
                <div>
                    <div class="text-right">
                        <SolidButton
                            type="submit"
                            disabled="{registerDisabled}">
                            Register
                        </SolidButton>
                    </div>
                </div>
            </form>
        </div>

        <div class="w-full md:w-1/2 px-4">
            <form
                on:submit="{createUserRegistered}"
                class="bg-white shadow-lg rounded p-4 md:p-6 mb-4"
                name="createAccount">
                <h2
                    class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                    md:leading-tight">
                    Create an Account (optional)
                </h2>

                <div class="mb-4">
                    <label
                        class="block text-gray-700 text-sm font-bold mb-2"
                        for="yourName2">
                        Name
                    </label>
                    <input
                        bind:value="{userName}"
                        placeholder="Enter your name"
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
                        id="yourName2"
                        name="yourName2"
                        required />
                </div>

                <div class="mb-4">
                    <label
                        class="block text-gray-700 text-sm font-bold mb-2"
                        for="yourEmail">
                        Email
                    </label>
                    <input
                        bind:value="{userEmail}"
                        placeholder="Enter your email"
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
                        id="yourEmail"
                        name="yourEmail"
                        type="email"
                        required />
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
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
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
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
                        id="yourPassword2"
                        name="yourPassword2"
                        type="password"
                        required />
                </div>

                <div>
                    <div class="text-right">
                        <SolidButton type="submit" disabled="{createDisabled}">
                            Create
                        </SolidButton>
                    </div>
                </div>
            </form>
        </div>
    </div>
</PageLayout>
