<script>
    import Navaid from 'navaid'
    import { onDestroy } from 'svelte'
    import Notifications from './components/Notifications.svelte'
    import UserIcon from './components/icons/UserIcon.svelte'
    import HollowButton from './components/HollowButton.svelte'

    import Landing from './pages/Landing.svelte'
    import Storyboards from './pages/Storyboards.svelte'
    import Storyboard from './pages/Storyboard.svelte'
    import Register from './pages/Register.svelte'
    import Login from './pages/Login.svelte'
    import ResetPassword from './pages/ResetPassword.svelte'
    import UserProfile from './pages/UserProfile.svelte'
    import VerifyAccount from './pages/VerifyAccount.svelte'
    import Admin from './pages/Admin.svelte'
    import { user } from './stores.js'

    const footerLinkClasses =
        'no-underline text-orange-500 hover:text-orange-800'

    let notifications

    let activeUser

    const unsubscribe = user.subscribe(w => {
        activeUser = w
    })

    let currentPage = {
        route: Landing,
        params: {},
    }

    const router = Navaid('/')
        .on('/', () => {
            currentPage = {
                route: Landing,
                params: {},
            }
        })
        .on('/register/:storyboardId?', params => {
            currentPage = {
                route: Register,
                params,
            }
        })
        .on('/login/:storyboardId?', params => {
            currentPage = {
                route: Login,
                params,
            }
        })
        .on('/reset-password/:resetId', params => {
            currentPage = {
                route: ResetPassword,
                params,
            }
        })
        .on('/verify-account/:verifyId', params => {
            currentPage = {
                route: VerifyAccount,
                params,
            }
        })
        .on('/user-profile', params => {
            currentPage = {
                route: UserProfile,
                params,
            }
        })
        .on('/storyboards', () => {
            currentPage = {
                route: Storyboards,
                params: {},
            }
        })
        .on('/storyboard/:storyboardId', params => {
            currentPage = {
                route: Storyboard,
                params,
            }
        })
        .on('/admin', () => {
            currentPage = {
                route: Admin,
                params: {},
            }
        })
        .listen()

    function logoutUser() {
        fetch('/api/auth/logout', {
            method: 'POST',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
            },
        })
            .then(function() {
                user.delete()
                router.route('/', true)
            })
            .catch(function(error) {
                notifications.danger(
                    'Error encountered attempting to logout user',
                )
            })
    }

    onDestroy(router.unlisten)
</script>

<style>
    :global(.nav-logo) {
        max-height: 3.75rem;
    }
    :global(.text-exo-red) {
        color: #a70b0f;
    }
    :global(.border-exo-red) {
        border-color: #a70b0f;
    }
    :global(.bg-exo-red) {
        background-color: #a70b0f;
    }
    :global(.text-exo-grey) {
        color: #2c2d2c;
    }
    :global(.bg-exo-grey) {
        background-color: #2c2d2c;
    }
</style>

<Notifications bind:this="{notifications}" />

<nav
    class="flex items-center justify-between flex-wrap bg-exo-grey px-6 py-2"
    role="navigation"
    aria-label="main navigation">
    <div class="flex items-center flex-shrink-0 mr-6">
        <a href="/">
            <img
                src="/img/exothermic-logo.png"
                alt="Exothermic"
                class="nav-logo" />
        </a>
    </div>
    {#if activeUser.name}
        <div class="text-right mt-4 md:mt-0">
            <span class="font-bold mr-2 text-xl text-white">
                <UserIcon />
                <a href="/user-profile">{activeUser.name}</a>
            </span>
            <HollowButton
                color="orange"
                href="/storyboards"
                additionalClasses="mr-2">
                My Storyboards
            </HollowButton>
            {#if !activeUser.type || activeUser.type === 'GUEST'}
                <HollowButton
                    color="orange"
                    href="/register"
                    additionalClasses="mr-2">
                    Create Account
                </HollowButton>
                <HollowButton href="/login">Login</HollowButton>
            {:else}
                {#if activeUser.type === 'ADMIN'}
                    <HollowButton
                        color="purple"
                        href="/admin"
                        additionalClasses="mr-2">
                        Admin
                    </HollowButton>
                {/if}
                <HollowButton color="red" onClick="{logoutUser}">
                    Logout
                </HollowButton>
            {/if}
        </div>
    {:else}
        <div class="text-right mt-4 md:mt-0">
            <HollowButton
                color="orange"
                href="/register"
                additionalClasses="mr-2">
                Create Account
            </HollowButton>
            <HollowButton href="/login">Login</HollowButton>
        </div>
    {/if}
</nav>

<svelte:component
    this="{currentPage.route}"
    {...currentPage.params}
    {notifications}
    {router} />

<footer class="p-6 text-center">
    <a
        href="https://github.com/StevenWeathers/exothermic-story-mapping"
        class="{footerLinkClasses}">
        Exothermic
    </a>
    by
    <a href="http://stevenweathers.com" class="{footerLinkClasses}">
        Steven Weathers
    </a>
    . The source code is licensed
    <a href="http://www.apache.org/licenses/" class="{footerLinkClasses}">
        Apache 2.0
    </a>
    .
    <br />
    Powered by
    <a href="https://svelte.dev/" class="{footerLinkClasses}">Svelte</a>
    and
    <a href="https://golang.org/" class="{footerLinkClasses}">Go</a>
</footer>
