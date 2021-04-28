<script>
    import SolidButton from './SolidButton.svelte'
    import CloseIcon from './icons/CloseIcon.svelte'
import HollowButton from './HollowButton.svelte'

    export let toggleStoryForm = () => {}
    export let updateContent = () => () => {}
    export let updateName = () => () => {}
    export let changeColor = () => () => {}
    export let updatePoints = () => () => {}
    export let updateClosed = () => () => {}
    export let deleteStory = () => () => {}

    export let story = {}

    const cardColors = [
        'red',
        'orange',
        'yellow',
        'green',
        'teal',
        'blue',
        'indigo',
        'purple',
        'pink',
    ]

    function handleStoryDelete() {
        deleteStory(story.id)()
        toggleStoryForm()
    }

    function markClosed() {
        updateClosed(story.id)(true)
        toggleStoryForm()
    }
    function markOpen() {
        updateClosed(story.id)(false)
        toggleStoryForm()
    }
</script>

<style>
    .story-red { @apply bg-red-400; }
    .story-red:hover { @apply bg-red-800; }
    .story-orange { @apply bg-orange-400; }
    .story-orange:hover { @apply bg-orange-800; }
    .story-yellow { @apply bg-yellow-400; }
    .story-yellow:hover { @apply bg-yellow-800; }
    .story-green { @apply bg-green-400; }
    .story-green:hover { @apply bg-green-800; }
    .story-teal { @apply bg-teal-400; }
    .story-teal:hover { @apply bg-teal-800; }
    .story-blue { @apply bg-blue-400; }
    .story-blue:hover { @apply bg-blue-800; }
    .story-indigo { @apply bg-indigo-400; }
    .story-indigo:hover { @apply bg-indigo-800; }
    .story-purple { @apply bg-purple-400; }
    .story-purple:hover { @apply bg-purple-800; }
    .story-pink { @apply bg-pink-400; }
    .story-pink:hover { @apply bg-pink-800; }
</style>

<div class="fixed inset-0 flex items-center z-40">
    <div class="fixed inset-0 bg-gray-900 opacity-75"></div>

    <div
        class="relative mx-4 md:mx-auto w-full md:w-2/3 z-50
        m-8">
        <div class="shadow-xl bg-white rounded-lg p-4 xl:p-6">
            <div class="flex justify-end mb-2">
                <button
                    aria-label="close"
                    on:click="{toggleStoryForm}"
                    class="text-gray-800">
                    <CloseIcon />
                </button>
            </div>

            <div class="flex w-full">
                <div class="w-3/4">
                    <div class="mx-4">
                        <div class="mb-4">
                            <label class="block text-sm font-bold mb-2" for="storyName">
                                Story Name
                            </label>
                            <input
                                class="bg-gray-200 border-gray-200 border-2
                                appearance-none rounded w-full py-2 px-3 text-gray-700
                                leading-tight focus:outline-none focus:bg-white
                                focus:border-purple-500"
                                id="storyName"
                                type="text"
                                on:change="{updateName(story.id)}"
                                value={story.name}
                                placeholder="Enter a story name e.g. Ricky Bobby"
                                name="storyName" />
                        </div>
                        <div class="mb-4">
                            <label class="block text-sm font-bold mb-2" for="storyDescription">
                                Story Content
                            </label>
                            <textarea
                                class="bg-gray-200 border-gray-200 border-2
                                appearance-none rounded w-full py-2 px-3 text-gray-700
                                leading-tight focus:outline-none focus:bg-white
                                focus:border-purple-500"
                                placeholder="Enter story content"
                                on:change="{updateContent(story.id)}"
                                value="{story.content}"></textarea>
                        </div>
                        <!-- <div class="mb-4">
                            <div class="font-bold text-lg">Discussion</div>
                            <div>
                                {#if story.comments}
                                    {#each story.comments as comment}
                                        <div>{comment.comment}</div>
                                    {/each}
                                {/if}
                            </div>
                        </div> -->
                    </div>
                </div>
                <div class="w-1/4">
                    <div class="mx-4">
                        <div class="mb-4">
                            <label class="block text-sm font-bold mb-2" for="storyPoints">
                                Story Points
                            </label>
                            <input
                                class="bg-gray-200 border-gray-200 border-2
                                appearance-none rounded w-full py-2 px-3 text-gray-700
                                leading-tight focus:outline-none focus:bg-white
                                focus:border-purple-500"
                                id="storyPoints"
                                type="number"
                                min="0"
                                max="999"
                                bind:value="{story.points}"
                                on:change="{updatePoints(story.id)}"
                                placeholder="Enter story points e.g. 1, 2, 3, 5, 8"
                                name="storyPoints" />
                        </div>
                        <div class="mb-2">
                            <div class="font-bold">Storycard Color</div>
                            <div>
                                {#each cardColors as color}
                                    <button
                                        on:click="{changeColor(story.id, color)}"
                                        class="p-4 mr-2 mb-2
                                        story-{color}
                                        border-2 border-solid
                                        {story.color === color ? `border-${color}-800` : "border-transparent"}"></button>
                                {/each}
                            </div>
                        </div>
                        <div class="mb-4">
                            {#if !story.closed}
                                <HollowButton color="orange" onClick={markClosed}>Mark story as Closed</HollowButton>
                            {:else}
                                <HollowButton color="green" onClick={markOpen}>Reopen story</HollowButton>
                            {/if}
                        </div>
                        <div class="text-right">
                            <HollowButton color="red" onClick={handleStoryDelete}>Delete Story</HollowButton>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
