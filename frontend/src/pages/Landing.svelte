<script>
    import CheckIcon from '../components/icons/CheckIcon.svelte'

    import { user } from '../stores.js'
</script>

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
                        href="{$user.id ? '/storyboards' : '/register'}">
                        Create a Storyboard
                    </a>
                </div>
            </div>
            <div class="w-full md:w-1/2 px-4">
                <img
                    class="w-full"
                    src="https://user-images.githubusercontent.com/846933/77712629-b8933500-6faa-11ea-9b9f-b64a1f648f98.png"
                    alt="Preview of Exothermic, Agile Story Mapping UI
                    experience" />
            </div>
        </div>
    </div>
</section>

<section class="bg-white">
    <div class="container mx-auto px-4 py-6 lg:py-10">
        <div class="text-center mb-6">
            <h2 class="text-2xl font-bold">Features</h2>
        </div>
        <div class="flex text-center mb-8">
            <div class="w-1/3">
                <CheckIcon />
                No limit on number of Users in a session
            </div>
            <div class="w-1/3">
                <CheckIcon />
                Ability to add goals (rows), columns and stories with common
                data points to a storyboard.
            </div>
            <div class="w-1/3">
                <CheckIcon />
                Open Source, feel free to
                <a
                    href="https://github.com/StevenWeathers/exothermic-story-mapping"
                    class="no-underline text-orange-500 hover:text-orange-800">
                    contribute and suggest features
                </a>
            </div>
        </div>
        <div class="flex text-center">
            <div class="w-full">
                <CheckIcon />
                No fees &amp; Ad free
            </div>
        </div>
    </div>
</section>
