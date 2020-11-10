<script>
    import PageLayout from '../components/PageLayout.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import UserRegisterForm from '../components/UserRegisterForm.svelte'
    import { user } from '../stores.js'
    import { validateName, validatePasswords } from '../validationUtils.js'
    import { appRoutes } from '../config'

    export let xfetch
    export let router
    export let eventTag
    export let notifications
    export let storyboardId

    const guestsAllowed = appConfig.AllowGuests
    const registrationAllowed = appConfig.AllowRegistration

    let userName = $user.name || ''

    $: targetPage = storyboardId
        ? `${appRoutes.storyboard}/${storyboardId}`
        : appRoutes.storyboards

    function createUserGuest(e) {
        e.preventDefault()
        const body = {
            userName,
        }
        const validName = validateName(userName)

        let noFormErrors = true

        if (!validName.valid) {
            noFormErrors = false
            notifications.danger(validName.error, 1500)
        }

        if (noFormErrors) {
            xfetch('/api/user', { body })
                .then(res => res.json())
                .then(function(newUser) {
                    user.create({
                        id: newUser.id,
                        name: newUser.name,
                        type: newUser.type,
                    })

                    eventTag('register_guest', 'engagement', 'success', () => {
                        router.route(targetPage, true)
                    })
                })
                .catch(function(error) {
                    notifications.danger(
                        'Error encountered registering user as guest',
                    )
                    eventTag('register_guest', 'engagement', 'failure')
                })
        }
    }

    function createUserRegistered(
        userName,
        userEmail,
        userPassword1,
        userPassword2,
    ) {
        const body = {
            userName,
            userEmail,
            userPassword1,
            userPassword2,
        }

        xfetch('/api/register', { body })
            .then(res => res.json())
            .then(function(newUser) {
                user.create({
                    id: newUser.id,
                    name: newUser.name,
                    email: newUser.email,
                    type: newUser.type,
                })

                eventTag(
                    'register_account',
                    'engagement',
                    'success',
                    () => {
                        router.route(targetPage, true)
                    },
                )
            })
            .catch(function(error) {
                notifications.danger('Error encountered creating user')
                eventTag('register_account', 'engagement', 'failure')
            })
    }

    $: registerDisabled = userName === ''
</script>

<PageLayout>
    <div class="text-center px-2 mb-4">
        <h1 class="text-3xl md:text-4xl font-bold">
            Register to Create Storyboards
        </h1>
    </div>
    <div class="flex flex-wrap">
        {#if !$user.id && guestsAllowed && registrationAllowed}
            <div class="w-full md:w-1/2 px-4">
                <form
                    on:submit="{createUserGuest}"
                    class="bg-white shadow-lg rounded p-4 md:p-6 mb-4"
                    name="registerGuest">
                    <h2
                        class="font-bold text-xl md:text-2xl b-4 mb-2 md:mb-6
                        md:leading-tight text-center">
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
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-orange-500"
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
        {/if}

        {#if registrationAllowed}
            <div class="w-full md:w-1/2 px-4">
                <div class="bg-white shadow-lg rounded p-4 md:p-6 mb-4">
                    <h2
                        class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                        md:leading-tight text-center">
                        Create an Account
                        <span class="text-gray-500">(optional)</span>
                    </h2>

                    <UserRegisterForm
                        guestUsersName="{userName}"
                        handleSubmit="{createUserRegistered}"
                        notifications />
                </div>
            </div>
        {:else}
            <div class="w-full md:w-1/2 px-4">
                <h2
                    class="font-bold text-2xl md:text-3xl md:leading-tight
                    text-center">
                    Registration is disabled.
                </h2>
            </div>
        {/if}
    </div>
</PageLayout>
