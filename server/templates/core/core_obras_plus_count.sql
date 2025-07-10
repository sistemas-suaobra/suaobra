select count(1) cnt
from core.core_obras_plus cop

left join main.obra_note n
  on cop.id = n.obra_id
  
where 1=1
  and city = '{city}'
  and size >= {sizeMin}
  and size <= {sizeMax}
  and {statusCond}
  and {neighborhoodCond}
  and {filterCond}