<style>
body {
	background: #eee;
	font-family: Roboto, sans-serif;
}
</style>

<style scoped>
main {
	min-width: 80%;
	display: flex;
	flex-direction: column;
	align-items: center;
}
#logo {
	margin: 8px 0 24px;
}
form {
	width: 600px;
	display: flex;
	align-items: center;
}

#result {
	text-align: center;
	font-size: 200%;
}

article {
	background: white;
	width: 100%;
	margin-top: -64px;
	padding: 120px 16px 16px;
	display: flex;
	flex-direction: column;
	align-items: center;
}
section {
	width: 960px;
	max-width: 100%;
	display: flex;
	align-items: center;
	justify-content: space-between;
}
article section:nth-child(even) {
	flex-flow: row-reverse;
}
section h1 {
	margin: 0;
	padding: 0;
	transition: font-size .3s ease;
}
section img {
	width: 500px;
	height: 500px;
	transition: opacity .3s, width .3s, height .3s;
}

footer {
	width: 100%;
	font-size: 90%;
	text-align: center;
	background: white;
	color: #aaa;
}

footer > a {
	color: #aaa !important;
	text-decoration: underline !important;
}

@media screen and (max-width: 960px) {
	article {
		padding: 120px 32px 16px;
	}
	section {
		justify-content: flex-end;
		position: relative;
		text-align: center;
		height: 500px;
	}
	section h1 {
		width: 100%;
		text-align: left;
	}
	article section:nth-child(even) h1 {
		text-align: right;
	}
	section img {
		position: absolute;
		opacity: .5;
	}
}

@media screen and (max-width: 600px) {
	main {
		display: block;
	}
	#logo {
		display: flex;
		justify-content: center;
		align-items: center;
		margin: 0 0 16px;
	}
	#card {
		margin: 8px;
	}
	form {
		display: block;
		width: auto;
		text-align: right;
	}
	article {
		margin-top: -80px;
	}
	section {
		height: 230px;
	}
	section h1 {
		font-size: 30px;
	}
	section img {
		width: 300px;
		height: 300px;
		opacity: .2;
	}
}

.v-enter-active, .v-leave-active {
	transition: all .3s ease;
}
.v-enter, .v-leave-to {
	transform: translateY(-24px);
	opacity: 0;
}
</style>

<template>
	<main>
		<div id=logo>
			<img alt="i1l.io" src=~/assets/logo.svg>
		</div>

		<md-card id=card>
			<md-card-content>
				<form @submit.prevent=onsubmit>
					<md-field>
						<md-input ref=input type=url v-model=url placeholder=http:// required />
					</md-field>
					<md-button class="md-raised md-primary" type=submit>SHORTEN</md-button>
				</form>

				<transition>
					<md-field v-if="shorten != null">
						<md-input ref=result id=result @click=copy :value=shorten readonly @copy.native=copied />
					</md-field>
				</transition>
			</md-card-content>
		</md-card>

		<article>
			<section>
				<h1 class="md-display-2">Easy to Reading</h1>
				<img src="~/assets/easy-to-reading.svg">
			</section>
			<section>
				<h1 class="md-display-2">Easy to Writing</h1>
				<img src="~/assets/easy-to-writing.svg">
			</section>
			<section>
				<h1 class="md-display-2">Everything is Great</h1>
				<img src="~/assets/everything-is-great.svg">
			</section>
		</article>

		<footer>
			MIT License (c)2018- <a href="https://blanktar.jp" target=_blank>MacRat</a>
		</footer>

		<md-dialog-alert
			:md-active="errorMessage != null"
			@md-closed="errorMessage = null"
			md-title="failed to shorten!"
			md-content=errorMessage
			/>

		<md-snackbar :md-active.sync=copyMessage :md-duration=3000>
			<span>copied to clipboard</span>
			<md-button class="md-primary" @click="copyMessage = false">OK</md-button>
		</md-snackbar>
	</main>
</template>

<script>
import Vue from 'vue';
import {MdButton, MdField, MdCard, MdSnackbar} from 'vue-material/dist/components';
import 'vue-material/dist/vue-material.min.css';
import 'vue-material/dist/theme/default.css';
import axios from 'axios';

Vue.use(MdButton);
Vue.use(MdField);
Vue.use(MdCard);
Vue.use(MdSnackbar);

export default {
	data() {
		return {
			url: 'http://',
			shorten: null,
			errorMessage: null,
			copyMessage: false,
		};
	},
	mounted() {
		this.$refs.input.$el.focus();
		this.$refs.input.$el.selectionStart = 10;
	},
	methods: {
		onsubmit() {
			axios({
				method: 'post',
				url: '/',
				headers: {'Content-Type': 'text/plain'},
				data: this.url,
				responseType: 'json',
			}).then(x => {
				if (x.data.error != null) {
					this.errorMessage = e;
				} else {
					this.shorten = x.data.info.short_url;
				}
			}).catch(e => this.errorMessage = e)
		},
		copy() {
			this.$refs.result.$el.select();
			document.execCommand('copy');
		},
		copied() {
			this.copyMessage = true;
		},
	},
}
</script>
