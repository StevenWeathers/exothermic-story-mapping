<script>
    import Countries from '../components/Countries.svelte'
    import CheckIcon from '../components/icons/CheckIcon.svelte'
    import GithubIcon from '../components/icons/Github.svelte'
    import { appRoutes } from '../config'

    import { user } from '../stores.js'

    export let xfetch
    export let eventTag

    const { ShowActiveCountries } = appConfig
</script>

<style>
    .feature-link {
        @apply no-underline;
        @apply text-blue-600;
    }
    .feature-link:hover {
        @apply text-blue-900;
    }
</style>

<section class="bg-exo-red text-exo-grey">
    <div class="container mx-auto px-4 py-6 lg:py-10">
        <div class="flex flex-wrap items-center -mx-4">
            <div class="w-full md:w-1/2 mb-4 lg:mb-0 px-4 text-white">
                <h1 class="mb-2 lg:mb-4 text-4xl font-bold leading-tight">
                    Exothermic is an Open Source Agile Story Mapping app
                </h1>

                <ul class="text-lg lg:text-2xl mb-4">
                    <li class="py-2 lg:py-4">
                        <CheckIcon />
                        Built on cool tech like Svelte, Go, and WebSockets
                    </li>
                    <li class="py-2 lg:py-4">
                        <CheckIcon />
                        Simple User Experience
                    </li>
                    <li class="py-2 lg:py-4">
                        <CheckIcon />
                        Works on all
                        <span class="font-semibold">modern browsers*</span>
                        , including mobile
                        <br />
                        <span class="text-sm">*sorry, not sorry IE</span>
                    </li>
                </ul>

                <div class="text-center mb-4 md:mb-0">
                    <a
                        class="w-full text-xl md:w-auto inline-block
                        no-underline bg-transparent bg-white
                        hover:bg-transparent hover:text-white font-semibold
                        text-exo-red py-4 px-10 border hover:border-white
                        border-transparent rounded"
                        href="{$user.id ? appRoutes.storyboards : appRoutes.register}">
                        Create a Storyboard
                    </a>
                </div>
            </div>
            <div class="w-full md:w-1/2 px-4">
                <img
                    class="w-full"
                    src="https://user-images.githubusercontent.com/846933/116486097-cfb1ef80-a85a-11eb-9769-96a9b5737ad9.png"
                    alt="Preview of Exothermic, Agile Story Mapping UI
                    experience" />
            </div>
        </div>
    </div>
</section>

<section class="bg-white">
    <div class="container mx-auto px-4 py-6 lg:py-10">
        <div class="text-center mb-6">
            <h2 class="text-4xl font-bold">Features</h2>
        </div>

        <div class="flex text-center mb-8">
            <div class="w-1/2">
                <h3 class="text-2xl font-bold">Open Source</h3>
                <p class="px-2">
                    Check out the <GithubIcon />
                    <a
                        href="https://github.com/StevenWeathers/exothermic-story-mapping"
                        class="feature-link">
                        repository
                    </a>
                    to request or contribute enhancements, locale
                    translations, and bug fixes or to
                    <a
                        href="https://github.com/StevenWeathers/exothermic-story-mapping#donations"
                        class="feature-link">
                        Donate
                    </a>
                </p>
            </div>
            <div class="w-1/2">
                <h3 class="text-2xl font-bold">
                    Hosted service or self-hosted
                </h3>
                <p class="px-2">
                    <a href="https://exothermic.dev" class="feature-link">
                        Exothermic.dev
                    </a>
                    is a free hosted service (SaaS) without Ads, alternatively
                    you can
                    <a
                        href="https://github.com/StevenWeathers/exothermic-story-mapping#running-in-production"
                        class="feature-link">
                        run it on your own server
                    </a>
                </p>
            </div>
        </div>

        <div class="flex text-center mb-8">
            <div class="w-1/2">
                <h3 class="text-2xl font-bold">Organizations and Teams</h3>
                <p class="px-2">
                    Create organizations with departments and teams at each
                    level to associate storyboard's to a team for quick access.
                    Just a small team?, no problem! Stand alone teams can be
                    created without the need for organizations.
                </p>
            </div>
            <div class="w-1/2">
                <h3 class="text-2xl font-bold">Feature rich story mapping</h3>
                <p class="px-2">
                    Start by creating goals with columns that can be named and
                    assigned persona's.
                    <br />
                    Add stories with details, commenting, card colors, and
                    more...
                </p>
            </div>
        </div>

        <div class="flex text-center mb-8">
            <div class="w-1/2">
                <h3 class="text-2xl font-bold">Fast</h3>
                <p class="px-2">
                    Exothermic is fast and small (under 20mb single executable)
                    thanks to
                    <a href="https://golang.org/" class="feature-link">Go</a>
                    .
                    <br />
                    <a href="https://svelte.dev/" class="feature-link">
                        Svelte
                    </a>
                    and Websockets provide a reactive web app experience.
                    <br />
                    <span class="font-bold">
                        Over 30 active users in a storyboard tested!
                    </span>
                </p>
            </div>
            <div class="w-1/2">
                <h3 class="text-2xl font-bold">Secure</h3>
                <p class="px-2">
                    Every storyboard has a randomly generated, not incremetal ID
                    which keeps strangers from joining your storyboard without
                    the share link.
                    <br />
                    HTTPS, backend only cookie, and salted/hashed password
                    storage keep your authenticated sessions secure.
                </p>
            </div>
        </div>
    </div>
</section>

{#if ShowActiveCountries}
    <Countries {xfetch} {eventTag} />
{/if}

<section class="mb-8">
    <div class="container mx-auto px-4 py-2 lg:py-4">
        <div class="flex text-center">
            <div class="w-full">
                <h4 class="font-bold text-xl">
                    Check out
                    <a href="https://thunderdome.dev">
                        <img
                            src="/img/thunderdome-logo.png"
                            alt="Thunderdome logo"
                            title="Thunderdome - Open Source Agile Planning
                            Poker tool"
                            class="inline-block mx-2" />
                    </a>
                    an open source agile planning poker tool brought to you by
                    yours truly!
                </h4>
            </div>
        </div>
    </div>
</section>
