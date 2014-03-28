#
# Copyright 2014 Nathaniel Cook
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

class Fitting(object):
    """
    Fittings are plugins to morgoth that expose behaviors

    Fittings can be used to insert data into morgoth via any format needed.
    Or they can be used to retrieve the data from morgoth in any desired format.

    This class is the base class for all Fittings

    """
    def __init__(self):
        pass

    @classmethod
    def from_conf(cls, conf):
        """
        Create a fitting from the given conf

        @param conf: a conf object
        """
        raise NotImplementedError("%s.from_conf is not implemented" % cls.__name__)

    def start(self):
        """ Start collecting data """
        raise NotImplementedError("%s.start is not implemented" % self.__class__.__name__)

    def stop(self):
        """ Stop the collection of data """
        raise NotImplementedError("%s.stop is not implemented" % self.__class__.__name__)
