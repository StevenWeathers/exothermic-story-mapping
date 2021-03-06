<script>
    import PageLayout from '../components/PageLayout.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import { user } from '../stores.js'
    import { appRoutes } from '../config'
    import { _, setupI18n } from '../i18n'

    export let xfetch
    export let router
    export let eventTag
    export let notifications
    export let storyboardId

    const { AllowRegistration, AuthMethod } = appConfig

    let userEmail = ''
    let userPassword = ''

    let userResetEmail = ''
    let forgotPassword = false

    $: targetPage = storyboardId
        ? `${appRoutes.storyboard}/${storyboardId}`
        : appRoutes.storyboards

    function authUser(e) {
        e.preventDefault()
        const body = {
            userEmail,
            userPassword,
        }

        xfetch('/api/auth', { body })
            .then(res => res.json())
            .then(function(newUser) {
                user.create({
                    id: newUser.id,
                    name: newUser.name,
                    email: newUser.email,
                    type: newUser.type,
                    locale: newUser.locale,
                })

                eventTag('login', 'engagement', 'success', () => {
                    setupI18n({
                        withLocale: newUser.locale,
                    })
                    router.route(targetPage, true)
                })
            })
            .catch(function(error) {
                notifications.danger(
                    'Error encountered attempting to authenticate user',
                )
                eventTag('login', 'engagement', 'failure')
            })
    }

    function toggleForgotPassword() {
        forgotPassword = !forgotPassword
        eventTag(
            'forgot_password_toggle',
            'engagement',
            `forgot: ${forgotPassword}`,
        )
    }

    function sendPasswordReset(e) {
        e.preventDefault()
        const body = {
            userEmail: userResetEmail,
        }

        xfetch('/api/auth/forgot-password', { body })
            .then(function() {
                notifications.success(
                    `
                    Password reset instructions sent to ${userResetEmail}.
                `,
                    2000,
                )
                forgotPassword = !forgotPassword
                eventTag('forgot_password', 'engagement', 'success')
            })
            .catch(function(error) {
                notifications.danger(
                    'Error encountered attempting to send password reset',
                )
                eventTag('forgot_password', 'engagement', 'failure')
            })
    }

    $: loginDisabled = userEmail === '' || userPassword === ''
    $: resetDisabled = userResetEmail === ''
</script>

<svelte:head>
    <title>Login | Exothermic</title>
</svelte:head>

<PageLayout>
    <div class="flex justify-center">
        <div class="w-full md:w-1/2 lg:w-1/3">
            {#if !forgotPassword}
                <form
                    on:submit="{authUser}"
                    class="bg-white shadow-lg rounded p-6 mb-4"
                    name="authUser">
                    <div
                        class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                        md:leading-tight text-center">
                        Login
                    </div>
                    {#if storyboardId && AllowRegistration}
                        <div
                            class="font-bold text-m md:text-l mb-2 md:mb-6
                            md:leading-tight text-center">
                            or
                            <a
                                href="{appRoutes.register}/{storyboardId}"
                                class="font-bold text-blue-500
                                hover:text-blue-800">
                                Register
                            </a>
                            to join the Storyboard
                        </div>
                    {/if}
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
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-orange-500"
                            id="yourEmail"
                            name="yourEmail"
                            type="email"
                            required />
                    </div>

                    <div class="mb-4">
                        <label
                            class="block text-gray-700 text-sm font-bold mb-2"
                            for="yourPassword">
                            Password
                        </label>
                        <input
                            bind:value="{userPassword}"
                            placeholder="Enter your password"
                            class="bg-gray-200 border-gray-200 border-2
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-orange-500"
                            id="yourPassword"
                            name="yourPassword"
                            type="password"
                            required />
                    </div>

                    <div class="text-right">
                        {#if AuthMethod === 'normal'}
                            <button
                                type="button"
                                class="inline-block align-baseline font-bold
                                text-sm text-blue-500 hover:text-blue-800 mr-4"
                                on:click="{toggleForgotPassword}">
                                Forgot Password?
                            </button>
                        {/if}
                        <SolidButton type="submit" disabled="{loginDisabled}">
                            Login
                        </SolidButton>
                    </div>
                </form>
            {/if}

            {#if forgotPassword}
                <form
                    on:submit="{sendPasswordReset}"
                    class="bg-white shadow-lg rounded p-6 mb-4"
                    name="resetPassword">
                    <div
                        class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                        md:leading-tight text-center">
                        Forgot Password
                    </div>
                    <div class="mb-4">
                        <label
                            class="block text-gray-700 text-sm font-bold mb-2"
                            for="yourResetEmail">
                            Email
                        </label>
                        <input
                            bind:value="{userResetEmail}"
                            placeholder="Enter your email"
                            class="bg-gray-200 border-gray-200 border-2
                            appearance-none rounded w-full py-2 px-3
                            text-gray-700 leading-tight focus:outline-none
                            focus:bg-white focus:border-orange-500"
                            id="yourResetEmail"
                            name="yourResetEmail"
                            type="email"
                            required />
                    </div>

                    <div class="text-right">
                        <button
                            type="button"
                            class="inline-block align-baseline font-bold text-sm
                            text-blue-500 hover:text-blue-800 mr-4"
                            on:click="{toggleForgotPassword}">
                            Cancel
                        </button>
                        <SolidButton type="submit" disabled="{resetDisabled}">
                            Send Reset Email
                        </SolidButton>
                    </div>
                </form>
            {/if}
        </div>
    </div>
</PageLayout>
