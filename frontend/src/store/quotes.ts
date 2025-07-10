
export interface MotivationalQuote {
  codigo: number;
  autor: string;
  frase: string
}

const text = `
codigo	autor	frase
1	Nick Gleason	Sucesso é a combinação de fracassos, erros, começos errados, confusão, e da determinação de continuar tentando mesmo assim.
2	Charles Robert Buxton	Você nunca encontrará tempo para nada. Se você quer tempo, você deve criá-lo.
3	Vince Lombardi	A diferença entre uma pessoa bem-sucedida e as outras não é a falta de força ou conhecimentos, mas sim de vontade.
4	Jeb Blount	Prospectar é a base, o coração de todo esforço de vendas de sucesso!
5	Jeb Blount	Ao ignorar a força da prospecção, muitos vendedores e organizações de vendas considerados competentes têm um desempenho abaixo do esperado.
6	Jeb Blount	Em vendas, nos negócios e na vida, há somente três coisas que você pode controlar: suas ações, suas reações e sua mentalidade.
7	Jeb Blount	Em vez de ficar chorando por coisas que estão fora do seu controle, foque sua energia no que você pode controlar – sua atitude, escolhas, emoções, objetivos, ambições, sonhos, desejos, disciplina
8	Jeb Blount	Quanto mais você prospectar, mais sortudo vai ficar.
9	Jeb Blount	A maioria dos vendedores nunca tem sorte porque faz apenas o mínimo de prospecções exigidas para passar raspando, e quando eles começam a prospectar, esperam milagres instantâneos.
10	Jeb Blount	Conheça seus números: Use sempre um CRM
11	Jeb Blount	Vendas são e sempre foram governadas por números porque, em vendas, a fórmula para o sucesso é uma simples fórmula matemática: o que (qualidade) você coloca no pipeline e quanto (quantidade) determina o que você tira do Funil de vendas
12	Jeb Blount	Você não pode se iludir e ter sucesso ao mesmo tempo.
13	Jeb Blount	Quando você escolhe a ilusão em vez da realidade, está mentindo para si mesmo, além de rebaixar seus padrões e seu desempenho.
14	Jeb Blount	Há quatro objetivos na prospecção: marcar uma reunião, qualificar informação, fechar uma venda, construir familiaridade.
15	Jeb Blount	Desenvolver um objetivo definido o torna efetivo porque, a cada ligação de prospecção, email, contato por mídia social, evento, você sabe exatamente o que perguntar ao seu cliente potencial, para lher dar uma razão estimulante para aceitar o seu pedido.
16	Jeb Blount	Mensagens de prospecção não são complexas.
17	Jeb Blount	Cuidado para não complicar demais as coisas. Sua mensagem de prospecção é feita para um único propósito: convencer rapidamente seu cliente potencial a ceder o tempo dele a você.
18	Jeb Blount	Peça o que você quer. O elemento mais importante de qualquer contato de prospecção é pedir.
19	Jeb Blount	Realmente não importa o que você diz ou faz. Se você não consegue pedir de maneira direta que seu cliente potencial tome providência, tudo o mais não passa de discurso acadêmico
20	Jeb Blount	Há somente uma técnica que realmente funciona para conseguir o que quer de um contato de prospecção. Pedir.
21	Jeb Blount	Seja confiante.
22	Jeb Blount	Quando os vendedores demonstram confiança e pedem com assertividade o que querem, clientes potenciais dizem sim em grande parte da vezes.
23	Jeb Blount	Cale a boca: A parte mais difícil de pedir uma venda é aprender a pedir e calar a boca. Depois que você perguntou o que quer aguarde o cliente potencial responder.
24	Jeb Blount	Voce precisa equilibrar suas acoes de prospeccao: (a) Chamadas Frias; (b) E-mail; (c) Whats app; (d) Mídias Sociais; (e) Telefone; (f) Presencial; (g) Networking; (h) Indicações.
25	Jeb Blount	Vender é interromper. Você precisa interromper o dia do outro para vender. E a ligação fria não é o mais difícil: o desconforto é interromper. Você não pode controlar a resposta das pessoas à sua interrupção, pois isso incomoda. Alguns clientes serão mais receptivos, outros não. Interromper o cliente é a base de um funil de vendas cheio.
26	Jeb Blount	Você não pode precisar desesperadamente das vendas. Quanto mais você precisar de algo, menor a probabilidade de conseguir aquilo
27	Jeb Blount	Conheça seus números. Você nunca vai chegar ao máximo de desempenho se não conhecer seus números e usá-los para fazer ajustes e render mais.
28	Jeb Blount	A maioria dos vendedores não registra e desconhece seus números (fuga da realidade?). Não acompanhar números é rebaixar padrões de desempenho.
29	Jeb Blount	Respeite a regra dos 30 dias. A prospecção feita nos últimos 30 dias repercutirá nos próximos 90 dias.
30	Jeb Blount	Vendas é estatística. Você precisa substituir as oportunidades que vão se perder naturalmente. Se você fecha 20% dos seus negócios, toda vez que você converte um cliente, você precisa colocar mais cinco no seu Funil de Vendas para manter seu ritmo.
31	Jeb Blount	Conheça sua taxa de conversão para saber o esforço que precisa todo novo dia. Fechar uma venda significa começar tudo de novo.
32	Jeb Blount	Defina seus objetivos para os contatos. O que você pretende e precisa fazer?
33	Jeb Blount	Defina seu plano de ataque: defina o perfil de cliente ideal; analise seu produto com relação à concorrência; tenha os diferenciais e seu apelo muito claros
34	Jeb Blount	Desenvolva boas perguntas para qualificar potenciais clientes; quanto mais familiaridade um cliente potencial tem com você e com sua marca, mais probabilidade de aceitar e retornar suas ligações, abrir seus e-mails, responder mensagens
35	Jeb Blount	Trabalhe seu funil de vendas. Clientes do topo: converter. Clientes entre o meio e o topo: construir familiaridade e monitorar janela de vendas. Clientes no meio: implantar campanhas e ações para antecipar janelas de compra. Clientes entre a base e o meio: identificar a janela de compra. Clientes da base: coletar informações sobre o cliente. Comece pelo topo e marque reuniões impactantes.
36	Jeb Blount	Como sair da baixa de vendas: A única maneira de se recuperar é prospectar. Defina metas diárias e aja todos os dias. Sucesso em vendas é uma questão de sistemática e não de magia
37	Jeb Blount	Destrave-se. Você não conseguirá fazer toda a prospecção de um mês em alguns dias. Não adie. Recupere o tempo perdido em até 48h. Não espere até ficar perfeito
38	Jeb Blount	Otimize sua agenda. Organize seu dia em blocos. Concentre-se e fuja da armadilha da multitarefa
39	Jeb Blount	Prepare-se ao máximo para render muito nas horas de prospecção direta. Use as horas de menos energia e concentração para se preparar  
40	Lucas Pelisson	Mantenha seu Funil de Vendas organizado. Voce vai me agradecer depois!
41	Jeb Blount	Os contatos com seus clientes é seu maior ativo! Isso nao é uma obrigacao ou chatisse, é a vida de seus negocios!!
42	Jeb Blount	O vendedor profissional deve: manter a integridade e precisão da base de dados; atualizar a base de dados em tempo real e realizar anotações detalhadas sobre prospecções
43	Jeb Blount	Trabalhe as indicações. Clientes satisfeitos: crie um processo disciplinado e sistemático para pedir indicações a estas pessoas
44	Jeb Blount	Capriche no contato com o cliente. O cliente não pode sentir medo, fraqueza ou falta de confiança. Você deve passar entusiasmo e confiança na sua mensagem (sem exagero).
45	Jeb Blount	Fale numa postura de confiança, mesmo quando não estiver presencialmente com a pessoa. Vista-se bem e fique bem com sua aparência física. Vá direto ao ponto, seja claro sobre suas intenções
46	Jeb Blount	Entenda as emoções do seu cliente. A emoção é um componente fundamental da tomada de decisão do seu cliente. Oferecer funcionalidades geralmente não funciona por conta disso. Ofereça valor emocional, de insight ou tangível (lógico).
47	Jeb Blount	Faça a ponte entre as emoções do cliente e seu discurso.
48	Jeb Blount	Simplesmente faça. Peça com confiança. Cale-se. Prepare-se para lidar com desvios por reflexo, dispensas e objeções. Vá ao ponto de modo direto, rápido e conciso.
49	Jeb Blount	Por que você precisa se calar. Se o cliente estiver muito propenso comprar, você vai chegar mais rápido ao “sim”. Se o cliente não quiser comprar de jeito nenhum, você vai chegar mais rápido ao “não”. Se o cliente estiver propenso a fazer uma objeção, ele fará.
50	Jeb Blount	Você precisa de um roteiro para lidar com objeções. Com roteiros, você gerencia a tensão do momento de receber uma rejeição
51	Roy Bartell	Muitas Pessoas pensam que “vender” é o mesmo que “falar”. Mas os vendedores mais eficazes sabem que ouvir é a parte mais importante do seu trabalho.
52	Lucas Pelisson	Você é o responsável pelo seu sucesso ou pelo seu fracasso.
53	Desconhecido	Seus hábitos diários são quem assinam seus cheques de comissões!
54	Salomão	Não sejas sábio a teus próprios olhos; tema a Deus e afaste-se do mal.
55	Salomão	O ingênuo acredita em tudo que ouve, o prudente examina seus passos com cuidado. O sábio é cauteloso e evita o mal, mas o tolo é impetuoso e irresponsável.
56	Salomão	Quando não há conselhos os planos se dispersam, mas havendo muitos conselheiros eles se firmam
57	Salomão	Lança o teu pão sobre as águas, porque depois de muitos dias o acharás. Reparte com sete, e ainda até com oito, porque não sabes que mal haverá sobre a terra
58	Salomão	Plante de manhã a sua semente, e mesmo ao entardecer não deixe as suas mãos ficarem à toa, pois você não sabe o que acontecerá, se esta ou aquela produzirá, ou se as duas serão igualmente boas.
59	Salomão	As mãos preguiçosas empobrecem o homem, porém as mãos diligentes lhe trazem riqueza.
60	Salomão	O preguiçoso deseja e nada consegue, mas os desejos do diligente são amplamente satisfeitos.
61	Salomão	O preguiçoso morre de tanto desejar e de nunca pôr as mãos no trabalho.
62	Jesus Cristo	Quem quiser tornar-se importante entre vocês deverá ser servo.
63	Salmos 112:5	Feliz é o homem que empresta com generosidade e que com honestidade conduz os seus negócios.
64	Salomão	O dinheiro ganho com desonestidade diminuirá, mas quem o ajunta aos poucos terá cada vez mais.
65	Salomão	Os planos bem elaborados levam à fartura; mas o apressado sempre acaba na miséria
66	Salomão	Consagre a Deus tudo o que você faz, e os seus planos serão bem-sucedidos
`

export const GetQuotes = () => {

  const lines = text.trim().split('\n')

  let quotes : MotivationalQuote[] = []
  for (let i = 0; i < lines.length; i++) {
    if(i === 0) continue // header line
    const values = lines[i].trim().split('\t')
    quotes.push({
      codigo: parseInt(values[0]),
      autor: values[1].trim(),
      frase: values[2].trim(),
    })
  }

  return quotes
}