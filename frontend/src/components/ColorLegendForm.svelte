<script>
    import SolidButton from './SolidButton.svelte'
    import CloseIcon from './icons/CloseIcon.svelte'

    export let handleLegendRevision = () => {}
    export let toggleEditLegend = () => {}

    export let colorLegend = []

    function handleSubmit(event) {
        event.preventDefault()

        handleLegendRevision(colorLegend)
        toggleEditLegend()
    }
</script>

<style>
    .colorcard-gray { @apply bg-gray-400; }
    .colorcard-red { @apply bg-red-400; }
    .colorcard-orange { @apply bg-orange-400; }
    .colorcard-yellow { @apply bg-yellow-400; }
    .colorcard-green { @apply bg-green-400; }
    .colorcard-teal { @apply bg-teal-400; }
    .colorcard-blue { @apply bg-blue-400; }
    .colorcard-indigo { @apply bg-indigo-400; }
    .colorcard-purple { @apply bg-purple-400; }
    .colorcard-pink { @apply bg-pink-400; }
</style>

<div class="fixed inset-0 flex items-center z-40">
    <div class="fixed inset-0 bg-gray-900 opacity-75"></div>

    <div
        class="relative mx-4 md:mx-auto w-full md:w-2/3 lg:w-3/5 xl:w-1/3 z-50
        m-8">
        <div class="shadow-xl bg-white rounded-lg p-4 xl:p-6">
            <div class="flex justify-end mb-2">
                <button
                    aria-label="close"
                    on:click="{toggleEditLegend}"
                    class="text-gray-800">
                    <CloseIcon />
                </button>
            </div>

            <form on:submit="{handleSubmit}" name="colorLegend">
                <div class="mb-4">
                    {#each colorLegend as color, i}
                        <div class="mb-1 flex">
                            <span class="p-4 inline-block colorcard-{color.color}"></span>
                            <input
                                class="bg-gray-200 border-gray-200 border-2
                                appearance-none rounded flex-grow ml-2 py-2 px-3 text-gray-700
                                leading-tight focus:outline-none focus:bg-white
                                focus:border-purple-500"
                                type="text"
                                bind:value={colorLegend[i].legend}
                                placeholder="Enter a color legend"
                                name="legend-{color.color}"
                                />
                        </div>
                    {/each}
                </div>
                <div class="text-right">
                    <div>
                        <SolidButton type="submit">Save</SolidButton>
                    </div>
                </div>
            </form>
        </div>
    </div>
</div>
