select
  c.id,
  c.team_id,
  c.canal,
  c.nome,
  case c.ativo when 1 then json('true') when 0 then json('false') end as ativo,
  c.status,
  c.taxa_por_min,
  c.tamanho_lote,
  c.limite_diario,
  c.ultimo_teste_em,
  c.ultimo_erro,
  c.created,
  c.updated,

  json_object(
    'id', c.id,
    'team_id', c.team_id,
    'canal', c.canal,
    'nome', c.nome,
    'ativo', case c.ativo when 1 then json('true') when 0 then json('false') end,
    'status', c.status,
    'taxa_por_min', c.taxa_por_min,
    'tamanho_lote', c.tamanho_lote,
    'limite_diario', c.limite_diario,
    'ultimo_teste_em', c.ultimo_teste_em,
    'ultimo_erro', c.ultimo_erro,
    'created', c.created,
    'updated', c.updated,

    'email', case
      when c.canal = 'EMAIL' and e.conexao is not null then json_object(
        'remetente_nome', e.remetente_nome,
        'remetente_email', e.remetente_email,
        'reply_to', e.reply_to,
        'smtp_host', e.smtp_host,
        'smtp_port', e.smtp_port,
        'smtp_usuario', e.smtp_usuario,
        'criptografia', e.criptografia,
        'taxa_por_min', c.taxa_por_min,
        'tamanho_lote', c.tamanho_lote,
        'limite_diario', c.limite_diario
      )
      else null
    end,

    'whatsapp', case
      when c.canal = 'WHATSAPP' and w.conexao is not null then json_object(
        'provider', w.provider,
        'api_base_url', w.api_base_url,
        'instancia_id', w.instancia_id,
        'numero_e164', w.numero_e164,
        'device_jid', w.device_jid,
        'conectado_em', w.conectado_em,
        'ultimo_qr_em', w.ultimo_qr_em,
        'taxa_por_min', c.taxa_por_min,
        'tamanho_lote', c.tamanho_lote,
        'limite_diario', c.limite_diario
      )
      else null
    end
  ) as conexao

from main.conexao c
left join main.conexao_email e on e.conexao = c.id
left join main.conexao_whatsapp w on w.conexao = c.id

where c.team_id = {:team_id}
order by c.created desc;