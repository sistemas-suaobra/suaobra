select distinct bairro
from core.core_obras_plus
where city = '{city}' and bairro is not null
order by 1