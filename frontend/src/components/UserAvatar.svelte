<script>
    import { _ } from '../i18n'
    const { PathPrefix } = appConfig

    export let avatarService = ''
    export let userId = ''
    export let avatar = ''
    export let width = 32
</script>

{#if avatarService === 'dicebear'}
    <img
        src="https://avatars.dicebear.com/api/{avatar}/{userId}.svg?w={width}"
        alt="{$_('avatarAltText')}" />
{:else if avatarService === 'gravatar'}
    <img
        src="https://gravatar.com/avatar/{userId}?s={width}&d={avatar}&r=g"
        alt="{$_('avatarAltText')}" />
{:else if avatarService === 'robohash'}
    <img
        src="https://robohash.org/{userId}.png?set={avatar}&size={width}x{width}"
        alt="{$_('avatarAltText')}" />
{:else if avatarService === 'govatar'}
    <img src="{PathPrefix}/avatar/{width}/{userId}/{avatar}" alt="{$_('avatarAltText')}" />
{:else if avatarService === 'goadorable'}
    <img src="{PathPrefix}/avatar/{width}/{userId}" alt="{$_('avatarAltText')}" />
{/if}
